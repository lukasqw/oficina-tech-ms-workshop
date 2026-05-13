package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// CancelConfirmedStockInput representa os dados de entrada para cancelar confirmação
type CancelConfirmedStockInput struct {
	ProductID string
	Quantity  int
}

// CancelConfirmedStockOutput representa os dados de saída após cancelar confirmação
type CancelConfirmedStockOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	UpdatedAt         string
}

// CancelConfirmedStockUseCase orquestra a operação de cancelamento de confirmação
type CancelConfirmedStockUseCase struct {
	inventoryRepo inventory.Repository
}

// NewCancelConfirmedStockUseCase cria uma nova instância do use case
func NewCancelConfirmedStockUseCase(inventoryRepo inventory.Repository) *CancelConfirmedStockUseCase {
	return &CancelConfirmedStockUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de cancelamento de confirmação
func (uc *CancelConfirmedStockUseCase) Execute(ctx context.Context, input CancelConfirmedStockInput) (*CancelConfirmedStockOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.cancel_confirmed")
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

	// Executar método CancelConfirmed da entidade
	if err := inv.CancelConfirmed(input.Quantity); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir alterações
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &CancelConfirmedStockOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
