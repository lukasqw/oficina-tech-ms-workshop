package usecases

import (
	"context"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

// GetProductByIDInput representa os dados de entrada para buscar um produto por ID
type GetProductByIDInput struct {
	ID string
}

// GetProductByIDOutput representa os dados de saída após buscar um produto
type GetProductByIDOutput struct {
	ID          string
	Name        string
	Description string
	Price       int
	ProductType string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GetProductByIDUseCase implementa o caso de uso de buscar produto por ID
type GetProductByIDUseCase struct {
	repository product.Repository
}

// NewGetProductByIDUseCase cria uma nova instância do caso de uso
func NewGetProductByIDUseCase(repository product.Repository) *GetProductByIDUseCase {
	return &GetProductByIDUseCase{
		repository: repository,
	}
}

// Execute executa o caso de uso de buscar produto por ID
func (uc *GetProductByIDUseCase) Execute(ctx context.Context, input GetProductByIDInput) (*GetProductByIDOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.get_product_by_id")
	defer span.End()

	// Validar UUID
	if err := utils.ValidateUUID(input.ID); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Busca o produto no repositório
	prod, err := uc.repository.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Converte a entidade de domínio para o output
	output := &GetProductByIDOutput{
		ID:          prod.ID(),
		Name:        prod.Name(),
		Description: prod.Description(),
		Price:       prod.Price(),
		ProductType: prod.ProductType().String(),
		CreatedAt:   prod.CreatedAt(),
		UpdatedAt:   prod.UpdatedAt(),
	}

	return output, nil
}

// ProductOutput representa um produto individual no output
type ProductOutput struct {
	ID          string
	Name        string
	Description string
	Price       int
	ProductType string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GetAllProductsOutput representa os dados de saída após buscar todos os produtos
type GetAllProductsOutput struct {
	Products []ProductOutput
}

// GetAllProductsUseCase implementa o caso de uso de buscar todos os produtos
type GetAllProductsUseCase struct {
	repository product.Repository
}

// NewGetAllProductsUseCase cria uma nova instância do caso de uso
func NewGetAllProductsUseCase(repository product.Repository) *GetAllProductsUseCase {
	return &GetAllProductsUseCase{
		repository: repository,
	}
}

// Execute executa o caso de uso de buscar todos os produtos
func (uc *GetAllProductsUseCase) Execute(ctx context.Context) (*GetAllProductsOutput, error) {
	ctx, span := observability.SpanUseCase(ctx, "inventory.get_all_products")
	defer span.End()

	// Busca todos os produtos no repositório
	products, err := uc.repository.FindAll(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Converte as entidades de domínio para o output
	output := &GetAllProductsOutput{
		Products: make([]ProductOutput, 0, len(products)),
	}

	for _, prod := range products {
		output.Products = append(output.Products, ProductOutput{
			ID:          prod.ID(),
			Name:        prod.Name(),
			Description: prod.Description(),
			Price:       prod.Price(),
			ProductType: prod.ProductType().String(),
			CreatedAt:   prod.CreatedAt(),
			UpdatedAt:   prod.UpdatedAt(),
		})
	}

	return output, nil
}
