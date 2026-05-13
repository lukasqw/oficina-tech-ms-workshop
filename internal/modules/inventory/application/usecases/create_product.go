package usecases

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

// CreateProductInput representa os dados de entrada para criar um produto
type CreateProductInput struct {
	Name        string
	Description string
	Price       int
	ProductType string
}

// CreateProductOutput representa os dados de saída após criar um produto
type CreateProductOutput struct {
	ID          string
	Name        string
	Description string
	Price       int
	ProductType string
	CreatedAt   string
	UpdatedAt   string
}

// CreateProductUseCase orquestra a criação de um novo produto
type CreateProductUseCase struct {
	productRepo   product.Repository
	inventoryRepo inventory.Repository
}

// NewCreateProductUseCase cria uma nova instância do caso de uso
func NewCreateProductUseCase(productRepo product.Repository, inventoryRepo inventory.Repository) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
	}
}

// Execute executa o caso de uso de criação de produto
func (uc *CreateProductUseCase) Execute(ctx context.Context, input CreateProductInput) (*CreateProductOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.create_product")
	defer span.End()

	// Criar value object ProductType
	productType, err := product.NewProductType(input.ProductType)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Criar entidade Product com validação
	newProduct, err := product.NewProduct(
		input.Name,
		input.Description,
		input.Price,
		productType,
	)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Salvar produto no repositório
	err = uc.productRepo.Save(ctx, newProduct)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Criar inventário automaticamente para o produto
	inv, err := inventory.NewInventory(newProduct.ID())
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Salvar inventário no repositório
	err = uc.inventoryRepo.Save(ctx, inv)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Retornar output
	return &CreateProductOutput{
		ID:          newProduct.ID(),
		Name:        newProduct.Name(),
		Description: newProduct.Description(),
		Price:       newProduct.Price(),
		ProductType: newProduct.ProductType().Value(),
		CreatedAt:   utils.FormatTimeRFC3339(newProduct.CreatedAt()),
		UpdatedAt:   utils.FormatTimeRFC3339(newProduct.UpdatedAt()),
	}, nil
}
