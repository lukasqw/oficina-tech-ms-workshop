package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

type UpdateServiceInput struct {
	ID          string
	Name        *string
	Description *string
	Price       *int
}

type UpdateServiceUseCase struct {
	repository service.Repository
}

func NewUpdateServiceUseCase(repository service.Repository) *UpdateServiceUseCase {
	return &UpdateServiceUseCase{
		repository: repository,
	}
}

func (uc *UpdateServiceUseCase) Execute(ctx context.Context, input UpdateServiceInput) error {
	ctx, span := observability.SpanUseCase(ctx, "service_catalog.update")
	defer span.End()

	// Validar UUID
	if err := utils.ValidateUUID(input.ID); err != nil {
		span.RecordError(service.ErrInvalidServiceID)
		return service.ErrInvalidServiceID
	}

	// Retrieve service
	svc, err := uc.repository.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Update name if provided
	if input.Name != nil {
		// Check for duplicate name if name is being changed
		if *input.Name != svc.Name() {
			exists, err := uc.repository.ExistsByName(ctx, *input.Name)
			if err != nil {
				span.RecordError(err)
				return err
			}
			if exists {
				span.RecordError(service.ErrDuplicateServiceName)
				return service.ErrDuplicateServiceName
			}
		}

		if err := svc.ChangeName(*input.Name); err != nil {
			span.RecordError(err)
			return err
		}
	}

	// Update description if provided
	if input.Description != nil {
		if err := svc.ChangeDescription(*input.Description); err != nil {
			span.RecordError(err)
			return err
		}
	}

	// Update price if provided
	if input.Price != nil {
		if err := svc.ChangePrice(*input.Price); err != nil {
			span.RecordError(err)
			return err
		}
	}

	// Save updated service
	if err := uc.repository.Save(ctx, svc); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
