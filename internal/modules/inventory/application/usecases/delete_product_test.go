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

func TestDeleteProductUseCase_Execute(t *testing.T) {
	validID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name          string
		input         DeleteProductInput
		mockProdRepo  *mockProductRepository
		mockInvRepo   *mockInventoryRepository
		expectedError error
	}{
		{
			name: "sucesso ao deletar produto com inventário",
			input: DeleteProductInput{
				ID: validID,
			},
			mockProdRepo: &mockProductRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			mockInvRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						10, // available
						0,  // reserved
						0,  // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "sucesso ao deletar produto sem inventário",
			input: DeleteProductInput{
				ID: validID,
			},
			mockProdRepo: &mockProductRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			mockInvRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ID inválido",
			input: DeleteProductInput{
				ID: "invalid-uuid",
			},
			mockProdRepo:  &mockProductRepository{},
			mockInvRepo:   &mockInventoryRepository{},
			expectedError: errors.New("invalid UUID format: invalid UUID length: 12"),
		},
		{
			name: "erro quando produto não é encontrado",
			input: DeleteProductInput{
				ID: validID,
			},
			mockProdRepo: &mockProductRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return product.ErrProductNotFound
				},
			},
			mockInvRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedError: product.ErrProductNotFound,
		},
		{
			name: "erro ao deletar inventário",
			input: DeleteProductInput{
				ID: validID,
			},
			mockProdRepo: &mockProductRepository{},
			mockInvRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						10, // available
						0,  // reserved
						0,  // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "erro ao deletar produto",
			input: DeleteProductInput{
				ID: validID,
			},
			mockProdRepo: &mockProductRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return errors.New("database error")
				},
			},
			mockInvRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewDeleteProductUseCase(tt.mockProdRepo, tt.mockInvRepo)
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
