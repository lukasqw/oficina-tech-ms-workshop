package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"

	"github.com/google/uuid"
)

func TestCreateProductUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		input         CreateProductInput
		mockProdRepo  *mockProductRepository
		mockInvRepo   *mockInventoryRepository
		expectedError error
	}{
		{
			name: "sucesso ao criar produto",
			input: CreateProductInput{
				Name:        "Óleo de Motor",
				Description: "Óleo sintético 5W30",
				Price:       5000,
				ProductType: "CONSUMABLE",
			},
			mockProdRepo: &mockProductRepository{
				saveFunc: func(_ context.Context, prod *product.Product) error {
					// Simula o SetID que seria feito pelo repositório
					_ = prod.SetID(uuid.New().String())
					return nil
				},
			},
			mockInvRepo: &mockInventoryRepository{
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer nome inválido (muito curto)",
			input: CreateProductInput{
				Name:        "A",
				Description: "Descrição",
				Price:       1000,
				ProductType: "SIMPLE",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: product.ErrInvalidProductName,
		},
		{
			name: "erro ao fornecer nome inválido (muito longo)",
			input: CreateProductInput{
				Name:        string(make([]byte, 201)),
				Description: "Descrição",
				Price:       1000,
				ProductType: "SIMPLE",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: product.ErrInvalidProductName,
		},
		{
			name: "erro ao fornecer preço inválido (zero)",
			input: CreateProductInput{
				Name:        "Produto Teste",
				Description: "Descrição",
				Price:       0,
				ProductType: "SIMPLE",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: product.ErrInvalidPrice,
		},
		{
			name: "erro ao fornecer preço negativo",
			input: CreateProductInput{
				Name:        "Produto Teste",
				Description: "Descrição",
				Price:       -100,
				ProductType: "SIMPLE",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: product.ErrInvalidPrice,
		},
		{
			name: "erro ao fornecer tipo de produto inválido",
			input: CreateProductInput{
				Name:        "Produto Teste",
				Description: "Descrição",
				Price:       1000,
				ProductType: "INVALID_TYPE",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: product.ErrInvalidProductType,
		},
		{
			name: "erro ao fornecer descrição muito longa",
			input: CreateProductInput{
				Name:        "Produto Teste",
				Description: string(make([]byte, 1001)),
				Price:       1000,
				ProductType: "SIMPLE",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: product.ErrInvalidDescription,
		},
		{
			name: "erro ao salvar produto no repositório",
			input: CreateProductInput{
				Name:        "Produto Teste",
				Description: "Descrição",
				Price:       1000,
				ProductType: "SIMPLE",
			},
			mockProdRepo: &mockProductRepository{
				saveFunc: func(_ context.Context, prod *product.Product) error {
					return errors.New("database error")
				},
			},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: errors.New("database error"),
		},
		{
			name: "erro ao salvar inventário no repositório",
			input: CreateProductInput{
				Name:        "Produto Teste",
				Description: "Descrição",
				Price:       1000,
				ProductType: "SIMPLE",
			},
			mockProdRepo: &mockProductRepository{
				saveFunc: func(_ context.Context, prod *product.Product) error {
					// Simula o SetID que seria feito pelo repositório
					_ = prod.SetID(uuid.New().String())
					return nil
				},
			},
			mockInvRepo: &mockInventoryRepository{
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewCreateProductUseCase(tt.mockProdRepo, tt.mockInvRepo)
			output, err := useCase.Execute(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("esperava erro %v, mas não obteve erro", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("esperava erro %v, obteve %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("não esperava erro, mas obteve: %v", err)
				return
			}

			if output == nil {
				t.Error("esperava output não nulo")
				return
			}

			if output.Name != tt.input.Name {
				t.Errorf("esperava Name %s, obteve %s", tt.input.Name, output.Name)
			}

			if output.Price != tt.input.Price {
				t.Errorf("esperava Price %d, obteve %d", tt.input.Price, output.Price)
			}

			if output.ProductType != tt.input.ProductType {
				t.Errorf("esperava ProductType %s, obteve %s", tt.input.ProductType, output.ProductType)
			}
		})
	}
}
