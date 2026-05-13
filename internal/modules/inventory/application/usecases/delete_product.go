package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

// DeleteProductInput representa os dados de entrada para deletar um produto
type DeleteProductInput struct {
	ID string
}

// DeleteProductUseCase orquestra a exclusão de um produto (soft delete)
type DeleteProductUseCase struct {
	productRepo   product.Repository
	inventoryRepo inventory.Repository
}

// NewDeleteProductUseCase cria uma nova instância do caso de uso
func NewDeleteProductUseCase(productRepo product.Repository, inventoryRepo inventory.Repository) *DeleteProductUseCase {
	return &DeleteProductUseCase{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de exclusão de produto
func (uc *DeleteProductUseCase) Execute(ctx context.Context, input DeleteProductInput) error {
	ctx, span := observability.SpanUseCase(ctx, "inventory.delete_product")
	defer span.End()

	// Validar UUID
	if err := utils.ValidateUUID(input.ID); err != nil {
		span.RecordError(err)
		return err
	}

	// Buscar inventário associado ao produto
	inv, err := uc.inventoryRepo.FindByProductID(ctx, input.ID)
	if err == nil && inv != nil {
		// Se existe inventário, deletar (soft delete)
		if err := uc.inventoryRepo.Delete(ctx, inv.ID()); err != nil {
			span.RecordError(err)
			return err
		}
	}
	// Se não existe inventário, continua normalmente

	// Deletar produto (soft delete)
	err = uc.productRepo.Delete(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
