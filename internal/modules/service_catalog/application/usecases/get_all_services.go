package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"time"
)

type ServiceOutput struct {
	ID          string
	Name        string
	Description string
	Price       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetAllServicesOutput struct {
	Services []ServiceOutput
}

type GetAllServicesUseCase struct {
	repository service.Repository
}

func NewGetAllServicesUseCase(repository service.Repository) *GetAllServicesUseCase {
	return &GetAllServicesUseCase{
		repository: repository,
	}
}

func (uc *GetAllServicesUseCase) Execute(ctx context.Context) (*GetAllServicesOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "service_catalog.get_all")
	defer span.End()

	// Retrieve all services from repository
	services, err := uc.repository.FindAll(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	output := &GetAllServicesOutput{
		Services: make([]ServiceOutput, 0, len(services)),
	}

	for _, svc := range services {
		output.Services = append(output.Services, ServiceOutput{
			ID:          svc.ID(),
			Name:        svc.Name(),
			Description: svc.Description(),
			Price:       svc.Price(),
			CreatedAt:   svc.CreatedAt(),
			UpdatedAt:   svc.UpdatedAt(),
		})
	}

	return output, nil
}
