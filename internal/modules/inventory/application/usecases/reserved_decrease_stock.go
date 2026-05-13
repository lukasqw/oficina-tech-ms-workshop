package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// ReservedDecreaseStockInput representa os dados de entrada para baixa de estoque reservado
type ReservedDecreaseStockInput struct {
	ProductID string
	Quantity  int
}

// ReservedDecreaseStockOutput representa os dados de saída após baixa de reservado
type ReservedDecreaseStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// ReservedDecreaseStockUseCase orquestra a operação de baixa de estoque reservado
type ReservedDecreaseStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewReservedDecreaseStockUseCase cria uma nova instância do use case
func NewReservedDecreaseStockUseCase(inventoryRepo inventory.Repository) *ReservedDecreaseStockUseCase {
	return &ReservedDecreaseStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de baixa de estoque reservado
func (uc *ReservedDecreaseStockUseCase) Execute(ctx context.Context, input ReservedDecreaseStockInput) (*ReservedDecreaseStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.reserved_decrease")
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

	// Executar método ReservedDecrease da entidade
	if err := inv.ReservedDecrease(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &ReservedDecreaseStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
