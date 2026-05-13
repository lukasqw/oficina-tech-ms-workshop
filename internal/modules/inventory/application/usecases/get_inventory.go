package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// GetInventoryInput representa os dados de entrada para buscar um estoque
type GetInventoryInput struct {
	ID string
}

// GetInventoryOutput representa os dados de saída após buscar um estoque
type GetInventoryOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	CreatedAt         string
	UpdatedAt         string
}

// GetInventoryUseCase orquestra a busca de um estoque por ID
type GetInventoryUseCase struct {
	inventoryRepo inventory.Repository
}

// NewGetInventoryUseCase cria uma nova instância do use case
func NewGetInventoryUseCase(inventoryRepo inventory.Repository) *GetInventoryUseCase {
	return &GetInventoryUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de busca de estoque
func (uc *GetInventoryUseCase) Execute(ctx context.Context, input GetInventoryInput) (*GetInventoryOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.get")
	defer span.End()

	// Validar ID
	if _, err := uuid.Parse(input.ID); err != nil {
		span.RecordError(inventory.ErrInvalidInventoryID)
		return nil, inventory.ErrInvalidInventoryID
	}

	// Buscar estoque por ID usando repositório
	inv, err := uc.inventoryRepo.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar dados completos do estoque
	return &GetInventoryOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		CreatedAt:         utils.FormatTimeRFC3339(inv.CreatedAt()),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
