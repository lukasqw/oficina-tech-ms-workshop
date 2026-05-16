package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
)

type ProcessSagaItemInput struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type ProcessSagaOperationInput struct {
	SagaID    string                   `json:"saga_id"`
	OrderID   string                   `json:"order_id"`
	Operation saga_operation.Operation `json:"operation"`
	Items     []ProcessSagaItemInput   `json:"items"`
	Now       func() time.Time         `json:"-"`
}

type ProcessSagaOperationOutput struct {
	EventName string
	Payload   []byte
	Succeeded bool
}

type orderInventoryOperationSucceeded struct {
	Event      string                   `json:"event"`
	SagaID     string                   `json:"saga_id"`
	OrderID    string                   `json:"order_id"`
	Operation  saga_operation.Operation `json:"operation"`
	OccurredAt string                   `json:"occurred_at"`
}

type orderInventoryOperationFailed struct {
	Event      string                   `json:"event"`
	SagaID     string                   `json:"saga_id"`
	OrderID    string                   `json:"order_id"`
	Operation  saga_operation.Operation `json:"operation"`
	Reason     string                   `json:"reason"`
	OccurredAt string                   `json:"occurred_at"`
}

type ProcessSagaOperationUseCase struct {
	sagaRepo              saga_operation.Repository
	reserveStock          *ReserveStockUseCase
	reservedDecreaseStock *ReservedDecreaseStockUseCase
	cancelReservedStock   *CancelReservedStockUseCase
	cancelConfirmedStock  *CancelConfirmedStockUseCase
}

func NewProcessSagaOperationUseCase(sagaRepo saga_operation.Repository, inventoryRepo inventory.Repository) *ProcessSagaOperationUseCase {
	return &ProcessSagaOperationUseCase{
		sagaRepo:              sagaRepo,
		reserveStock:          NewReserveStockUseCase(inventoryRepo),
		reservedDecreaseStock: NewReservedDecreaseStockUseCase(inventoryRepo),
		cancelReservedStock:   NewCancelReservedStockUseCase(inventoryRepo),
		cancelConfirmedStock:  NewCancelConfirmedStockUseCase(inventoryRepo),
	}
}

func (uc *ProcessSagaOperationUseCase) Execute(ctx context.Context, input ProcessSagaOperationInput) (*ProcessSagaOperationOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.process_saga_operation")
	defer span.End()

	existing, err := uc.sagaRepo.FindBySagaAndOperation(ctx, input.SagaID, input.Operation)
	if err == nil {
		if existing.Status() == saga_operation.StatusCompleted || existing.Status() == saga_operation.StatusFailed {
			return outputFromStoredPayload(existing.ResultPayload())
		}
		return nil, fmt.Errorf("saga operation already processing: saga_id=%s operation=%s", input.SagaID, input.Operation)
	}
	if err != saga_operation.ErrSagaOperationNotFound {
		span.RecordError(err)
		return nil, err
	}

	operation, err := saga_operation.NewSagaOperation(input.SagaID, input.OrderID, input.Operation)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if err := uc.sagaRepo.Save(ctx, operation); err != nil {
		span.RecordError(err)
		return nil, err
	}

	applied := make([]ProcessSagaItemInput, 0, len(input.Items))
	for _, item := range input.Items {
		if err := uc.apply(ctx, input.Operation, item); err != nil {
			span.RecordError(err)
			uc.compensateApplied(ctx, input.Operation, applied)
			output, payloadErr := uc.fail(ctx, operation, input, err.Error())
			if payloadErr != nil {
				span.RecordError(payloadErr)
				return nil, payloadErr
			}
			return output, nil
		}
		applied = append(applied, item)
	}

	return uc.complete(ctx, operation, input)
}

func (uc *ProcessSagaOperationUseCase) apply(ctx context.Context, operation saga_operation.Operation, item ProcessSagaItemInput) error {
	switch operation {
	case saga_operation.OperationReserve:
		_, err := uc.reserveStock.Execute(ctx, ReserveStockInput(item))
		return err
	case saga_operation.OperationReservedDecrease:
		_, err := uc.reservedDecreaseStock.Execute(ctx, ReservedDecreaseStockInput(item))
		return err
	case saga_operation.OperationCancelReserved:
		_, err := uc.cancelReservedStock.Execute(ctx, CancelReservedStockInput(item))
		return err
	case saga_operation.OperationCancelConfirmed:
		_, err := uc.cancelConfirmedStock.Execute(ctx, CancelConfirmedStockInput(item))
		return err
	default:
		return saga_operation.ErrInvalidOperation
	}
}

func (uc *ProcessSagaOperationUseCase) compensateApplied(ctx context.Context, operation saga_operation.Operation, applied []ProcessSagaItemInput) {
	for i := len(applied) - 1; i >= 0; i-- {
		_ = uc.apply(ctx, compensationFor(operation), applied[i])
	}
}

func compensationFor(operation saga_operation.Operation) saga_operation.Operation {
	switch operation {
	case saga_operation.OperationReserve:
		return saga_operation.OperationCancelReserved
	case saga_operation.OperationReservedDecrease:
		return saga_operation.OperationCancelConfirmed
	case saga_operation.OperationCancelReserved:
		return saga_operation.OperationReserve
	case saga_operation.OperationCancelConfirmed:
		return saga_operation.OperationReservedDecrease
	default:
		return operation
	}
}

func (uc *ProcessSagaOperationUseCase) complete(ctx context.Context, operation *saga_operation.SagaOperation, input ProcessSagaOperationInput) (*ProcessSagaOperationOutput, error) {
	payload, err := json.Marshal(orderInventoryOperationSucceeded{
		Event:      "OrderInventoryOperationSucceeded",
		SagaID:     input.SagaID,
		OrderID:    input.OrderID,
		Operation:  input.Operation,
		OccurredAt: occurredAt(input.Now),
	})
	if err != nil {
		return nil, err
	}

	operation.Complete(payload)
	if err := uc.sagaRepo.Save(ctx, operation); err != nil {
		return nil, err
	}

	return &ProcessSagaOperationOutput{
		EventName: "OrderInventoryOperationSucceeded",
		Payload:   payload,
		Succeeded: true,
	}, nil
}

func (uc *ProcessSagaOperationUseCase) fail(ctx context.Context, operation *saga_operation.SagaOperation, input ProcessSagaOperationInput, reason string) (*ProcessSagaOperationOutput, error) {
	payload, err := json.Marshal(orderInventoryOperationFailed{
		Event:      "OrderInventoryOperationFailed",
		SagaID:     input.SagaID,
		OrderID:    input.OrderID,
		Operation:  input.Operation,
		Reason:     reason,
		OccurredAt: occurredAt(input.Now),
	})
	if err != nil {
		return nil, err
	}

	operation.Fail(payload)
	if err := uc.sagaRepo.Save(ctx, operation); err != nil {
		return nil, err
	}

	return &ProcessSagaOperationOutput{
		EventName: "OrderInventoryOperationFailed",
		Payload:   payload,
		Succeeded: false,
	}, nil
}

func outputFromStoredPayload(payload []byte) (*ProcessSagaOperationOutput, error) {
	var envelope struct {
		Event string `json:"event"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, err
	}

	return &ProcessSagaOperationOutput{
		EventName: envelope.Event,
		Payload:   append([]byte(nil), payload...),
		Succeeded: envelope.Event == "OrderInventoryOperationSucceeded",
	}, nil
}

func occurredAt(now func() time.Time) string {
	if now == nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return now().UTC().Format(time.RFC3339)
}
