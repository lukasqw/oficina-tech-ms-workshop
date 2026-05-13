package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// CancelReservedStockInput representa os dados de entrada para cancelar reserva
type CancelReservedStockInput struct {
	ProductID string
	Quantity  int
}

// CancelReservedStockOutput representa os dados de saída após cancelar reserva
type CancelReservedStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// CancelReservedStockUseCase orquestra a operação de cancelamento de reserva
type CancelReservedStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewCancelReservedStockUseCase cria uma nova instância do use case
func NewCancelReservedStockUseCase(inventoryRepo inventory.Repository) *CancelReservedStockUseCase {
	return &CancelReservedStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de cancelamento de reserva
func (uc *CancelReservedStockUseCase) Execute(ctx context.Context, input CancelReservedStockInput) (*CancelReservedStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.cancel_reserved")
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

	// Executar método CancelReserved da entidade
	if err := inv.CancelReserved(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &CancelReservedStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
