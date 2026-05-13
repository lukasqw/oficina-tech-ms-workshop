package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"

	"github.com/google/uuid"
)

// DeleteInventoryInput representa os dados de entrada para deletar um estoque
type DeleteInventoryInput struct {
	ID string
}

// DeleteInventoryUseCase orquestra a deleção de um estoque
type DeleteInventoryUseCase struct {
	inventoryRepo inventory.Repository
}

// NewDeleteInventoryUseCase cria uma nova instância do use case
func NewDeleteInventoryUseCase(inventoryRepo inventory.Repository) *DeleteInventoryUseCase {
	return &DeleteInventoryUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de deleção de estoque
func (uc *DeleteInventoryUseCase) Execute(ctx context.Context, input DeleteInventoryInput) error {
	ctx, span := observability.SpanUseCase(ctx, "inventory.delete")
	defer span.End()

	// Validar ID
	if _, err := uuid.Parse(input.ID); err != nil {
		span.RecordError(inventory.ErrInvalidInventoryID)
		return inventory.ErrInvalidInventoryID
	}

	// Executar soft delete usando repositório
	if err := uc.inventoryRepo.Delete(ctx, input.ID); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
