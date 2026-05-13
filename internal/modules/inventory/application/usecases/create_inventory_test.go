package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"

	"github.com/google/uuid"
)

// Mock do repositório de product
type mockProductRepository struct {
	findByIDFunc func(ctx context.Context, id string) (*product.Product, error)
	saveFunc     func(ctx context.Context, prod *product.Product) error
	findAllFunc  func(ctx context.Context) ([]*product.Product, error)
	deleteFunc   func(ctx context.Context, id string) error
}

func (m *mockProductRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockProductRepository) Save(ctx context.Context, prod *product.Product) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, prod)
	}
	return nil
}

func (m *mockProductRepository) FindAll(ctx context.Context) ([]*product.Product, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockProductRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestCreateInventoryUseCase_Execute(t *testing.T) {
	validProductID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name          string
		input         CreateInventoryInput
		mockInvRepo   *mockInventoryRepository
		mockProdRepo  *mockProductRepository
		expectedError error
	}{
		{
			name: "sucesso ao criar inventário",
			input: CreateInventoryInput{
				ProductID: validProductID,
			},
			mockInvRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return false, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			mockProdRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Produto Teste",
						"Descrição",
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
			name: "erro ao fornecer ProductID inválido",
			input: CreateInventoryInput{
				ProductID: "invalid-uuid",
			},
			mockInvRepo:   &mockInventoryRepository{},
			mockProdRepo:  &mockProductRepository{},
			expectedError: inventory.ErrInvalidProductID,
		},
		{
			name: "erro quando produto não existe",
			input: CreateInventoryInput{
				ProductID: validProductID,
			},
			mockInvRepo: &mockInventoryRepository{},
			mockProdRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},
			expectedError: product.ErrProductNotFound,
		},
		{
			name: "erro quando produto já possui inventário",
			input: CreateInventoryInput{
				ProductID: validProductID,
			},
			mockInvRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return true, nil
				},
			},
			mockProdRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Produto Teste",
						"Descrição",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: inventory.ErrProductAlreadyHasInventory,
		},
		{
			name: "erro ao verificar existência de inventário",
			input: CreateInventoryInput{
				ProductID: validProductID,
			},
			mockInvRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return false, errors.New("database error")
				},
			},
			mockProdRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Produto Teste",
						"Descrição",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "erro ao salvar inventário",
			input: CreateInventoryInput{
				ProductID: validProductID,
			},
			mockInvRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return false, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return errors.New("database error")
				},
			},
			mockProdRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					productType, _ := product.NewProductType("SIMPLE")
					prod, _ := product.ReconstructProduct(
						id,
						"Produto Teste",
						"Descrição",
						1000,
						productType,
						now,
						now,
						nil,
					)
					return prod, nil
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewCreateInventoryUseCase(tt.mockInvRepo, tt.mockProdRepo)
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

			if output.ProductID != tt.input.ProductID {
				t.Errorf("esperava ProductID %s, obteve %s", tt.input.ProductID, output.ProductID)
			}

			if output.AvailableQuantity != 0 {
				t.Errorf("esperava AvailableQuantity 0, obteve %d", output.AvailableQuantity)
			}

			if output.ReservedQuantity != 0 {
				t.Errorf("esperava ReservedQuantity 0, obteve %d", output.ReservedQuantity)
			}

			if output.PendingQuantity != 0 {
				t.Errorf("esperava PendingQuantity 0, obteve %d", output.PendingQuantity)
			}
		})
	}
}
