package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"

	"github.com/google/uuid"
)

// CreateInventoryInput representa os dados de entrada para criar um estoque
type CreateInventoryInput struct {
	ProductID string
}

// CreateInventoryOutput representa os dados de saída após criar um estoque
type CreateInventoryOutput struct {
	ID                string
	ProductID         string
	AvailableQuantity int
	ReservedQuantity  int
	PendingQuantity   int
	CreatedAt         string
	UpdatedAt         string
}

// CreateInventoryUseCase orquestra a criação de um novo registro de estoque
type CreateInventoryUseCase struct {
	inventoryRepo inventory.Repository
	productRepo   product.Repository
}

// NewCreateInventoryUseCase cria uma nova instância do use case
func NewCreateInventoryUseCase(inventoryRepo inventory.Repository, productRepo product.Repository) *CreateInventoryUseCase {
	return &CreateInventoryUseCase{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
	}
}

// Execute executa o caso de uso de criação de estoque
func (uc *CreateInventoryUseCase) Execute(ctx context.Context, input CreateInventoryInput) (*CreateInventoryOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.create")
	defer span.End()

	// Validar ProductID
	if _, err := uuid.Parse(input.ProductID); err != nil {
		span.RecordError(inventory.ErrInvalidProductID)
		return nil, inventory.ErrInvalidProductID
	}

	// Verificar se o produto existe
	_, err := uc.productRepo.FindByID(ctx, input.ProductID)
	if err != nil {
		span.RecordError(product.ErrProductNotFound)
		return nil, product.ErrProductNotFound
	}

	// Verificar se o produto já possui estoque
	exists, err := uc.inventoryRepo.ExistsByProductID(ctx, input.ProductID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if exists {
		span.RecordError(inventory.ErrProductAlreadyHasInventory)
		return nil, inventory.ErrProductAlreadyHasInventory
	}

	// Criar nova entidade Inventory com quantidades zeradas
	inv, err := inventory.NewInventory(input.ProductID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Persistir usando repositório
	if err := uc.inventoryRepo.Save(ctx, inv); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &CreateInventoryOutput{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		CreatedAt:         utils.FormatTimeRFC3339(inv.CreatedAt()),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}, nil
}
