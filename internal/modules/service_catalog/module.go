package service_catalog

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/application/usecases"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/dto"
)

// ServiceCatalogModule is the facade interface for the service_catalog bounded context
// It exposes operations needed by other modules
type ServiceCatalogModule interface {
	GetServiceByID(ctx context.Context, id string) (*dto.ServiceDTO, error)
}

// serviceCatalogModuleImpl implements the ServiceCatalogModule interface
type serviceCatalogModuleImpl struct {
	getServiceUseCase *usecases.GetServiceByIDUseCase
}

// NewServiceCatalogModule creates a new instance of ServiceCatalogModule
func NewServiceCatalogModule(serviceRepo service.Repository) ServiceCatalogModule {
	return &serviceCatalogModuleImpl{
		getServiceUseCase: usecases.NewGetServiceByIDUseCase(serviceRepo),
	}
}

// GetServiceByID retrieves a service by its ID and returns it as a ServiceDTO
func (m *serviceCatalogModuleImpl) GetServiceByID(ctx context.Context, id string) (*dto.ServiceDTO, error) {
	// Execute the use case
	output, err := m.getServiceUseCase.Execute(ctx, usecases.GetServiceByIDInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	// Convert domain entity output to shared DTO
	return &dto.ServiceDTO{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Price:       output.Price,
	}, nil
}
