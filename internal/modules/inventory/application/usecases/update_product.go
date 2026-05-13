package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

// UpdateProductInput representa os dados de entrada para atualizar um produto
type UpdateProductInput struct {
	ID          string
	Name        *string
	Description *string
	Price       *int
	ProductType *string
}

// UpdateProductUseCase orquestra a atualização de um produto existente
type UpdateProductUseCase struct {
	repository product.Repository
}

// NewUpdateProductUseCase cria uma nova instância do caso de uso
func NewUpdateProductUseCase(repository product.Repository) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		repository: repository,
	}
}

// Execute executa o caso de uso de atualização de produto
func (uc *UpdateProductUseCase) Execute(ctx context.Context, input UpdateProductInput) error {
	ctx, span := observability.SpanUseCase(ctx, "inventory.update_product")
	defer span.End()

	// Validar UUID
	if err := utils.ValidateUUID(input.ID); err != nil {
		span.RecordError(err)
		return err
	}

	// Buscar produto existente
	existingProduct, err := uc.repository.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Aplicar mudanças apenas nos campos fornecidos
	if input.Name != nil {
		err = existingProduct.ChangeName(*input.Name)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	if input.Description != nil {
		err = existingProduct.ChangeDescription(*input.Description)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	if input.Price != nil {
		err = existingProduct.ChangePrice(*input.Price)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	if input.ProductType != nil {
		// Criar value object ProductType com validação
		productType, err := product.NewProductType(*input.ProductType)
		if err != nil {
			span.RecordError(err)
			return err
		}
		existingProduct.ChangeProductType(productType)
	}

	// Salvar produto atualizado
	err = uc.repository.Save(ctx, existingProduct)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
