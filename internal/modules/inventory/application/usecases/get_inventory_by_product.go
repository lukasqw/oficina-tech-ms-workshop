package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// GetInventoryByProductInput representa os dados de entrada para buscar um estoque por produto
type GetInventoryByProductInput struct {
	ProductID string
}

// GetInventoryByProductOutput representa os dados de saída após buscar um estoque por produto
type GetInventoryByProductOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	CreatedAt         string
	UpdatedAt         string
}

// GetInventoryByProductUseCase orquestra a busca de um estoque por ProductID
type GetInventoryByProductUseCase struct {
	inventoryRepo inventory.Repository
}

// NewGetInventoryByProductUseCase cria uma nova instância do use case
func NewGetInventoryByProductUseCase(inventoryRepo inventory.Repository) *GetInventoryByProductUseCase {
	return &GetInventoryByProductUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de busca de estoque por produto
func (uc *GetInventoryByProductUseCase) Execute(ctx context.Context, input GetInventoryByProductInput) (*GetInventoryByProductOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.get_by_product")
	defer span.End()

	// Validar ProductID
	if _, err := uuid.Parse(input.ProductID); err != nil {
		span.RecordError(inventory.ErrInvalidProductID)
		return nil, inventory.ErrInvalidProductID
	}

	// Buscar estoque por ProductID usando repositório
	inv, err := uc.inventoryRepo.FindByProductID(ctx, input.ProductID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar dados completos do estoque
	return &GetInventoryByProductOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		CreatedAt:         utils.FormatTimeRFC3339(inv.CreatedAt()),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
