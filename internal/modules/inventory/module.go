package inventory

import (
	"context"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/dto"
)

// InventoryModule is the facade interface for the inventory bounded context
// It exposes operations needed by other modules
type InventoryModule interface {
	GetProductByID(ctx context.Context, id string) (*dto.ProductDTO, error)
	ChangeStock(ctx context.Context, operation dto.StockOperationDTO) error
}

// inventoryModuleImpl implements the InventoryModule interface
type inventoryModuleImpl struct {
	getProductUseCase             *usecases.GetProductByIDUseCase
	reserveStockUseCase           *usecases.ReserveStockUseCase
	reservedDecreaseStockUseCase  *usecases.ReservedDecreaseStockUseCase
	cancelReservedStockUseCase    *usecases.CancelReservedStockUseCase
	cancelConfirmedStockUseCase   *usecases.CancelConfirmedStockUseCase
	availableDecreaseStockUseCase *usecases.AvailableDecreaseStockUseCase
	increaseStockUseCase          *usecases.IncreaseStockUseCase
}

// NewInventoryModule creates a new instance of InventoryModule
func NewInventoryModule(productRepo product.Repository, inventoryRepo inventory.Repository) InventoryModule {
	return &inventoryModuleImpl{
		getProductUseCase:             usecases.NewGetProductByIDUseCase(productRepo),
		reserveStockUseCase:           usecases.NewReserveStockUseCase(inventoryRepo),
		reservedDecreaseStockUseCase:  usecases.NewReservedDecreaseStockUseCase(inventoryRepo),
		cancelReservedStockUseCase:    usecases.NewCancelReservedStockUseCase(inventoryRepo),
		cancelConfirmedStockUseCase:   usecases.NewCancelConfirmedStockUseCase(inventoryRepo),
		availableDecreaseStockUseCase: usecases.NewAvailableDecreaseStockUseCase(inventoryRepo),
		increaseStockUseCase:          usecases.NewIncreaseStockUseCase(inventoryRepo),
	}
}

// GetProductByID retrieves a product by its ID and returns it as a ProductDTO
func (m *inventoryModuleImpl) GetProductByID(ctx context.Context, id string) (*dto.ProductDTO, error) {
	// Execute the use case
	output, err := m.getProductUseCase.Execute(ctx, usecases.GetProductByIDInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	// Convert domain entity output to shared DTO
	return &dto.ProductDTO{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Price:       output.Price,
		ProductType: output.ProductType,
	}, nil
}

// ChangeStock executes a stock operation based on the operation type
func (m *inventoryModuleImpl) ChangeStock(ctx context.Context, operation dto.StockOperationDTO) error {
	// Validate operation type
	switch operation.OperationType {
	case dto.StockOpReserve:
		_, err := m.reserveStockUseCase.Execute(ctx, usecases.ReserveStockInput{
			ProductID: operation.ProductID,
			Quantity:  operation.Quantity,
		})
		return err

	case dto.StockOpReservedDecrease:
		_, err := m.reservedDecreaseStockUseCase.Execute(ctx, usecases.ReservedDecreaseStockInput{
			ProductID: operation.ProductID,
			Quantity:  operation.Quantity,
		})
		return err

	case dto.StockOpCancelReserved:
		_, err := m.cancelReservedStockUseCase.Execute(ctx, usecases.CancelReservedStockInput{
			ProductID: operation.ProductID,
			Quantity:  operation.Quantity,
		})
		return err

	case dto.StockOpCancelConfirmed:
		_, err := m.cancelConfirmedStockUseCase.Execute(ctx, usecases.CancelConfirmedStockInput{
			ProductID: operation.ProductID,
			Quantity:  operation.Quantity,
		})
		return err

	case dto.StockOpAvailableDecrease:
		_, err := m.availableDecreaseStockUseCase.Execute(ctx, usecases.AvailableDecreaseStockInput{
			ProductID: operation.ProductID,
			Quantity:  operation.Quantity,
		})
		return err

	case dto.StockOpIncrease:
		_, err := m.increaseStockUseCase.Execute(ctx, usecases.IncreaseStockInput{
			ProductID: operation.ProductID,
			Quantity:  operation.Quantity,
		})
		return err

	default:
		return inventory.ErrInvalidOperation
	}
}
