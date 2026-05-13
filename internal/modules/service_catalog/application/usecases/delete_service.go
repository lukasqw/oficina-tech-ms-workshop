package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

type DeleteServiceInput struct {
	ID string
}

type DeleteServiceUseCase struct {
	repository service.Repository
}

func NewDeleteServiceUseCase(repository service.Repository) *DeleteServiceUseCase {
	return &DeleteServiceUseCase{
		repository: repository,
	}
}

func (uc *DeleteServiceUseCase) Execute(ctx context.Context, input DeleteServiceInput) error {
	ctx, span := observability.SpanUseCase(ctx, "service_catalog.delete")
	defer span.End()

	// Validar UUID
	if err := utils.ValidateUUID(input.ID); err != nil {
		span.RecordError(service.ErrInvalidServiceID)
		return service.ErrInvalidServiceID
	}

	// Validate that service exists before deleting
	_, err := uc.repository.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(service.ErrServiceNotFound)
		return service.ErrServiceNotFound
	}

	// Delete service from repository by ID
	if err := uc.repository.Delete(ctx, input.ID); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
