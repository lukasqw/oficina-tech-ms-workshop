package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
	"gorm.io/gorm"
)

type SagaOperationRepositoryImpl struct {
	db *gorm.DB
}

func NewSagaOperationRepository(db *gorm.DB) saga_operation.Repository {
	return &SagaOperationRepositoryImpl{db: db}
}

func (r *SagaOperationRepositoryImpl) Save(ctx context.Context, operation *saga_operation.SagaOperation) error {
	model, err := FromDomainSagaOperation(operation)
	if err != nil {
		return err
	}

	if operation.ID() == "" {
		model.ID = uuid.Must(uuid.NewV7())
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
		return operation.SetID(model.ID.String())
	}

	return r.db.WithContext(ctx).Save(model).Error
}

func (r *SagaOperationRepositoryImpl) FindBySagaAndOperation(ctx context.Context, sagaID string, operation saga_operation.Operation) (*saga_operation.SagaOperation, error) {
	uid, err := uuid.Parse(sagaID)
	if err != nil {
		return nil, saga_operation.ErrInvalidSagaID
	}

	var model SagaOperationModel
	err = r.db.WithContext(ctx).
		First(&model, "saga_id = ? AND operation = ?", uid, string(operation)).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, saga_operation.ErrSagaOperationNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}
