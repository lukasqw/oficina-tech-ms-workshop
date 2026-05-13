package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// ManualDecreaseStockInput representa os dados de entrada para baixa manual de estoque
type ManualDecreaseStockInput struct {
	ProductID string
	Quantity  int
}

// ManualDecreaseStockOutput representa os dados de saída após baixa manual
type ManualDecreaseStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// ManualDecreaseStockUseCase orquestra a operação de baixa manual de estoque
type ManualDecreaseStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewManualDecreaseStockUseCase cria uma nova instância do use case
func NewManualDecreaseStockUseCase(inventoryRepo inventory.Repository) *ManualDecreaseStockUseCase {
	return &ManualDecreaseStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de baixa manual de estoque
func (uc *ManualDecreaseStockUseCase) Execute(ctx context.Context, input ManualDecreaseStockInput) (*ManualDecreaseStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.manual_decrease")
	defer span.End()

	// Validar ProductID
	if _, err := uuid.Parse(input.ProductID); err != nil {
		span.RecordError(inventory.ErrInvalidProductID)
		return nil, inventory.ErrInvalidProductID
	}

	// Buscar estoque por ProductID
	inv, err := uc.inventoryRepo.FindByProductID(ctx, input.ProductID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Executar método ManualDecrease da entidade
	if err := inv.ManualDecrease(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &ManualDecreaseStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
