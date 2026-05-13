package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"time"
)

type GetServiceByIDInput struct {
	ID string
}

type GetServiceByIDOutput struct {
	ID          string
	Name        string
	Description string
	Price       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetServiceByIDUseCase struct {
	repository service.Repository
}

func NewGetServiceByIDUseCase(repository service.Repository) *GetServiceByIDUseCase {
	return &GetServiceByIDUseCase{
		repository: repository,
	}
}

func (uc *GetServiceByIDUseCase) Execute(ctx context.Context, input GetServiceByIDInput) (*GetServiceByIDOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "service_catalog.get_by_id")
	defer span.End()

	// Validar UUID
	if err := utils.ValidateUUID(input.ID); err != nil {
		span.RecordError(service.ErrInvalidServiceID)
		return nil, service.ErrInvalidServiceID
	}

	// Retrieve service from repository by ID
	svc, err := uc.repository.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &GetServiceByIDOutput{
		ID:          svc.ID(),
		Name:        svc.Name(),
		Description: svc.Description(),
		Price:       svc.Price(),
		CreatedAt:   svc.CreatedAt(),
		UpdatedAt:   svc.UpdatedAt(),
	}, nil
}
