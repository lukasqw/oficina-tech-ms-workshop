package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"

	"github.com/google/uuid"
)

func TestUpdateProductUseCase_Execute(t *testing.T) {
	validID := uuid.New().String()
	now := time.Now()

	newName := "Novo Nome"
	newDescription := "Nova Descrição"
	newPrice := 2000
	newProductType := "PARTS"

	tests := []struct {
		name          string
		input         UpdateProductInput
		mockRepo      *mockProductRepository
		expectedError error
	}{
		{
			name: "sucesso ao atualizar todos os campos",
			input: UpdateProductInput{
				ID:          validID,
				Name:        &newName,
				Description: &newDescription,
				Price:       &newPrice,
				ProductType: &newProductType,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Nome Antigo",
						"Descrição Antiga",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
				saveFunc: func(_ context.Context, prod *product.Product) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "sucesso ao atualizar apenas o nome",
			input: UpdateProductInput{
				ID:   validID,
				Name: &newName,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Nome Antigo",
						"Descrição Antiga",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
				saveFunc: func(_ context.Context, prod *product.Product) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ID inválido",
			input: UpdateProductInput{
				ID:   "invalid-uuid",
				Name: &newName,
			},
			mockRepo:      &mockProductRepository{},
			expectedError: errors.New("invalid UUID format: invalid UUID length: 12"),
		},
		{
			name: "erro quando produto não é encontrado",
			input: UpdateProductInput{
				ID:   validID,
				Name: &newName,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},
			expectedError: product.ErrProductNotFound,
		},
		{
			name: "erro ao fornecer nome inválido (muito curto)",
			input: UpdateProductInput{
				ID:   validID,
				Name: func() *string { s := "A"; return &s }(),
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Nome Antigo",
						"Descrição Antiga",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: product.ErrInvalidProductName,
		},
		{
			name: "erro ao fornecer preço inválido (zero)",
			input: UpdateProductInput{
				ID:    validID,
				Price: func() *int { p := 0; return &p }(),
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Nome Antigo",
						"Descrição Antiga",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: product.ErrInvalidPrice,
		},
		{
			name: "erro ao fornecer tipo de produto inválido",
			input: UpdateProductInput{
				ID:          validID,
				ProductType: func() *string { s := "INVALID"; return &s }(),
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Nome Antigo",
						"Descrição Antiga",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: product.ErrInvalidProductType,
		},
		{
			name: "erro ao salvar no repositório",
			input: UpdateProductInput{
				ID:   validID,
				Name: &newName,
			},
			mockRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Nome Antigo",
						"Descrição Antiga",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
				saveFunc: func(_ context.Context, prod *product.Product) error {
					return errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewUpdateProductUseCase(tt.mockRepo)
			err := useCase.Execute(context.Background(), tt.input)

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
			}
		})
	}
}
