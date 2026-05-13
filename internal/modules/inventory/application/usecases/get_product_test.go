package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"

	"github.com/google/uuid"
)

func TestGetProductByIDUseCase_Execute(t *testing.T) {
	validID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name          string
		input         GetProductByIDInput
		mockRepo      *mockProductRepository
		expectedError error
	}{
		{
			name: "sucesso ao buscar produto",
			input: GetProductByIDInput{
				ID: validID,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Produto Teste",
						"Descrição do produto",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ID inválido",
			input: GetProductByIDInput{
				ID: "invalid-uuid",
			},
			mockRepo:      &mockProductRepository{},
			expectedError: errors.New("invalid UUID format: invalid UUID length: 12"),
		},
		{
			name: "erro quando produto não é encontrado",
			input: GetProductByIDInput{
				ID: validID,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},
			expectedError: product.ErrProductNotFound,
		},
		{
			name: "erro ao buscar no repositório",
			input: GetProductByIDInput{
				ID: validID,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					return nil, errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewGetProductByIDUseCase(tt.mockRepo)
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

			if output.ID != tt.input.ID {
				t.Errorf("esperava ID %s, obteve %s", tt.input.ID, output.ID)
			}

			if output.Name != "Produto Teste" {
				t.Errorf("esperava Name 'Produto Teste', obteve %s", output.Name)
			}

			if output.Price != 1000 {
				t.Errorf("esperava Price 1000, obteve %d", output.Price)
			}
		})
	}
}

func TestGetAllProductsUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		mockRepo      *mockProductRepository
		expectedError error
		expectedCount int
	}{
		{
			name: "sucesso ao buscar todos os produtos",
			mockRepo: &mockProductRepository{
				findAllFunc: func(_ context.Context) ([]*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod1, _ := product.ReconstructProduct(
						uuid.New().String(),
						"Produto 1",
						"Descrição 1",
						1000,
						productType,
						now,
						now,
						nil,
					)
					prod2, _ := product.ReconstructProduct(
						uuid.New().String(),
						"Produto 2",
						"Descrição 2",
						2000,
						productType,
						now,
						now,
						nil,
					)
					return []*product.Product{prod1, prod2}, nil
				},
			},
			expectedError: nil,
			expectedCount: 2,
		},
		{
			name: "sucesso ao buscar lista vazia",
			mockRepo: &mockProductRepository{
				findAllFunc: func(_ context.Context) ([]*product.Product, error) {
					return []*product.Product{}, nil
				},
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name: "erro ao buscar no repositório",
			mockRepo: &mockProductRepository{
				findAllFunc: func(_ context.Context) ([]*product.Product, error) {
					return nil, errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewGetAllProductsUseCase(tt.mockRepo)
			output, err := useCase.Execute(context.Background())

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

			if len(output.Products) != tt.expectedCount {
				t.Errorf("esperava %d produtos, obteve %d", tt.expectedCount, len(output.Products))
			}
		})
	}
}
