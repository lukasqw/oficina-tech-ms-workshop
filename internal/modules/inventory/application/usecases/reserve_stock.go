package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// ReserveStockInput representa os dados de entrada para reservar estoque
type ReserveStockInput struct {
	ProductID string
	Quantity  int
}

// ReserveStockOutput representa os dados de saída após reservar estoque
type ReserveStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// ReserveStockUseCase orquestra a operação de reserva de estoque
type ReserveStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewReserveStockUseCase cria uma nova instância do use case
func NewReserveStockUseCase(inventoryRepo inventory.Repository) *ReserveStockUseCase {
	return &ReserveStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de reserva de estoque
func (uc *ReserveStockUseCase) Execute(ctx context.Context, input ReserveStockInput) (*ReserveStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.reserve")
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

	// Executar método Reserve da entidade
	if err := inv.Reserve(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &ReserveStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
