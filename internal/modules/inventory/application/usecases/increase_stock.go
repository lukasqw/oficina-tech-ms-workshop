package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// IncreaseStockInput representa os dados de entrada para aumentar estoque
type IncreaseStockInput struct {
	ProductID string
	Quantity  int
}

// IncreaseStockOutput representa os dados de saída após aumentar estoque
type IncreaseStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// IncreaseStockUseCase orquestra a operação de aumento de estoque
type IncreaseStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewIncreaseStockUseCase cria uma nova instância do use case
func NewIncreaseStockUseCase(inventoryRepo inventory.Repository) *IncreaseStockUseCase {
	return &IncreaseStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de aumento de estoque
func (uc *IncreaseStockUseCase) Execute(ctx context.Context, input IncreaseStockInput) (*IncreaseStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.increase")
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

	// Executar método Increase da entidade
	if err := inv.Increase(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &IncreaseStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
