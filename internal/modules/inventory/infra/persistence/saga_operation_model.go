package persistence

import (
	"time"

	"github.com/google/uuid"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
	"gorm.io/datatypes"
)

type SagaOperationModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	SagaID        uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_saga_operation"`
	OrderID       uuid.UUID      `gorm:"type:uuid;not null"`
	Operation     string         `gorm:"not null;size:40;uniqueIndex:idx_saga_operation"`
	Status        string         `gorm:"not null;size:40"`
	ResultPayload datatypes.JSON `gorm:"type:jsonb"`
	ProcessedAt   *time.Time
}

func (SagaOperationModel) TableName() string {
	return "saga_operations"
}

func (m *SagaOperationModel) ToDomain() (*saga_operation.SagaOperation, error) {
	return saga_operation.ReconstructSagaOperation(
		m.ID.String(),
		m.SagaID.String(),
		m.OrderID.String(),
		saga_operation.Operation(m.Operation),
		saga_operation.Status(m.Status),
		[]byte(m.ResultPayload),
		m.ProcessedAt,
	)
}

func FromDomainSagaOperation(operation *saga_operation.SagaOperation) (*SagaOperationModel, error) {
	var id uuid.UUID
	var err error
	if operation.ID() != "" {
		id, err = uuid.Parse(operation.ID())
		if err != nil {
			return nil, err
		}
	}

	sagaID, err := uuid.Parse(operation.SagaID())
	if err != nil {
		return nil, err
	}

	orderID, err := uuid.Parse(operation.OrderID())
	if err != nil {
		return nil, err
	}

	return &SagaOperationModel{
		ID:            id,
		SagaID:        sagaID,
		OrderID:       orderID,
		Operation:     string(operation.Operation()),
		Status:        string(operation.Status()),
		ResultPayload: datatypes.JSON(operation.ResultPayload()),
		ProcessedAt:   operation.ProcessedAt(),
	}, nil
}
