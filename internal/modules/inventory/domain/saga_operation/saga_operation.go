package saga_operation

import (
	"time"

	"github.com/google/uuid"
)

type Operation string

const (
	OperationReserve          Operation = "RESERVE"
	OperationReservedDecrease Operation = "RESERVED_DECREASE"
	OperationCancelReserved   Operation = "CANCEL_RESERVED"
	OperationCancelConfirmed  Operation = "CANCEL_CONFIRMED"
)

type Status string

const (
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
)

type SagaOperation struct {
	id            string
	sagaID        string
	orderID       string
	operation     Operation
	status        Status
	resultPayload []byte
	processedAt   *time.Time
}

func NewSagaOperation(sagaID, orderID string, operation Operation) (*SagaOperation, error) {
	if err := validateUUID(sagaID, ErrInvalidSagaID); err != nil {
		return nil, err
	}
	if err := validateUUID(orderID, ErrInvalidOrderID); err != nil {
		return nil, err
	}
	if !operation.Valid() {
		return nil, ErrInvalidOperation
	}

	return &SagaOperation{
		sagaID:    sagaID,
		orderID:   orderID,
		operation: operation,
		status:    StatusProcessing,
	}, nil
}

func ReconstructSagaOperation(id, sagaID, orderID string, operation Operation, status Status, resultPayload []byte, processedAt *time.Time) (*SagaOperation, error) {
	if err := validateUUID(id, ErrInvalidSagaOperationID); err != nil {
		return nil, err
	}
	if err := validateUUID(sagaID, ErrInvalidSagaID); err != nil {
		return nil, err
	}
	if err := validateUUID(orderID, ErrInvalidOrderID); err != nil {
		return nil, err
	}
	if !operation.Valid() {
		return nil, ErrInvalidOperation
	}
	if !status.Valid() {
		return nil, ErrInvalidStatus
	}

	return &SagaOperation{
		id:            id,
		sagaID:        sagaID,
		orderID:       orderID,
		operation:     operation,
		status:        status,
		resultPayload: resultPayload,
		processedAt:   processedAt,
	}, nil
}

func (o Operation) Valid() bool {
	switch o {
	case OperationReserve, OperationReservedDecrease, OperationCancelReserved, OperationCancelConfirmed:
		return true
	default:
		return false
	}
}

func (s Status) Valid() bool {
	switch s {
	case StatusProcessing, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

func (s *SagaOperation) ID() string {
	return s.id
}

func (s *SagaOperation) SagaID() string {
	return s.sagaID
}

func (s *SagaOperation) OrderID() string {
	return s.orderID
}

func (s *SagaOperation) Operation() Operation {
	return s.operation
}

func (s *SagaOperation) Status() Status {
	return s.status
}

func (s *SagaOperation) ResultPayload() []byte {
	return s.resultPayload
}

func (s *SagaOperation) ProcessedAt() *time.Time {
	return s.processedAt
}

func (s *SagaOperation) SetID(id string) error {
	if err := validateUUID(id, ErrInvalidSagaOperationID); err != nil {
		return err
	}
	s.id = id
	return nil
}

func (s *SagaOperation) Complete(resultPayload []byte) {
	now := time.Now()
	s.status = StatusCompleted
	s.resultPayload = append([]byte(nil), resultPayload...)
	s.processedAt = &now
}

func (s *SagaOperation) Fail(resultPayload []byte) {
	now := time.Now()
	s.status = StatusFailed
	s.resultPayload = append([]byte(nil), resultPayload...)
	s.processedAt = &now
}

func validateUUID(value string, err error) error {
	if _, parseErr := uuid.Parse(value); parseErr != nil {
		return err
	}
	return nil
}
