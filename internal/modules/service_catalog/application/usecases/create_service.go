package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"time"
)

type CreateServiceInput struct {
	Name        string
	Description string
	Price       int
}

type CreateServiceOutput struct {
	ID          string
	Name        string
	Description string
	Price       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateServiceUseCase struct {
	repository service.Repository
}

func NewCreateServiceUseCase(repository service.Repository) *CreateServiceUseCase {
	return &CreateServiceUseCase{
		repository: repository,
	}
}

func (uc *CreateServiceUseCase) Execute(ctx context.Context, input CreateServiceInput) (*CreateServiceOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "service_catalog.create")
	defer span.End()

	// Check for duplicate name
	exists, err := uc.repository.ExistsByName(ctx, input.Name)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if exists {
		span.RecordError(service.ErrDuplicateServiceName)
		return nil, service.ErrDuplicateServiceName
	}

	// Create Service entity
	svc, err := service.NewService(input.Name, input.Description, input.Price)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Save to repository
	if err := uc.repository.Save(ctx, svc); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &CreateServiceOutput{
		ID:          svc.ID(),
		Name:        svc.Name(),
		Description: svc.Description(),
		Price:       svc.Price(),
		CreatedAt:   svc.CreatedAt(),
		UpdatedAt:   svc.UpdatedAt(),
	}, nil
}
