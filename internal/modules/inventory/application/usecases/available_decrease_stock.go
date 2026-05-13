package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// AvailableDecreaseStockInput representa os dados de entrada para baixa de estoque disponível
type AvailableDecreaseStockInput struct {
	ProductID string
	Quantity  int
}

// AvailableDecreaseStockOutput representa os dados de saída após baixa de disponível
type AvailableDecreaseStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// AvailableDecreaseStockUseCase orquestra a operação de baixa de estoque disponível
type AvailableDecreaseStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewAvailableDecreaseStockUseCase cria uma nova instância do use case
func NewAvailableDecreaseStockUseCase(inventoryRepo inventory.Repository) *AvailableDecreaseStockUseCase {
	return &AvailableDecreaseStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de baixa de estoque disponível
func (uc *AvailableDecreaseStockUseCase) Execute(ctx context.Context, input AvailableDecreaseStockInput) (*AvailableDecreaseStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.available_decrease")
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

	// Executar método AvailableDecrease da entidade
	if err := inv.AvailableDecrease(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &AvailableDecreaseStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
