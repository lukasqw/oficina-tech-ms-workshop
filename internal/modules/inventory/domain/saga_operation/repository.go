package saga_operation

import "context"

type Repository interface {
	Save(ctx context.Context, operation *SagaOperation) error
	FindBySagaAndOperation(ctx context.Context, sagaID string, operation Operation) (*SagaOperation, error)
}
