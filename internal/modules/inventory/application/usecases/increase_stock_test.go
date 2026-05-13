package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"

	"github.com/google/uuid"
)

func TestIncreaseStockUseCase_Execute(t *testing.T) {
	validProductID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name          string
		input         IncreaseStockInput
		mockRepo      *mockInventoryRepository
		expectedError error
	}{
		{
			name: "sucesso ao aumentar estoque sem pendências",
			input: IncreaseStockInput{
				ProductID: validProductID,
				Quantity:  10,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						5, // available
						0, // reserved
						0, // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "sucesso ao aumentar estoque com pendências",
			input: IncreaseStockInput{
				ProductID: validProductID,
				Quantity:  10,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						5,  // available
						0,  // reserved
						15, // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ProductID inválido",
			input: IncreaseStockInput{
				ProductID: "invalid-uuid",
				Quantity:  10,
			},
			mockRepo:      &mockInventoryRepository{},
			expectedError: inventory.ErrInvalidProductID,
		},
		{
			name: "erro quando estoque não é encontrado",
			input: IncreaseStockInput{
				ProductID: validProductID,
				Quantity:  10,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedError: inventory.ErrInventoryNotFound,
		},
		{
			name: "erro ao fornecer quantidade inválida (zero)",
			input: IncreaseStockInput{
				ProductID: validProductID,
				Quantity:  0,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						5, // available
						0, // reserved
						0, // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
			},
			expectedError: inventory.ErrInvalidQuantity,
		},
		{
			name: "erro ao fornecer quantidade negativa",
			input: IncreaseStockInput{
				ProductID: validProductID,
				Quantity:  -10,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						5, // available
						0, // reserved
						0, // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
			},
			expectedError: inventory.ErrInvalidQuantity,
		},
		{
			name: "erro ao salvar no repositório",
			input: IncreaseStockInput{
				ProductID: validProductID,
				Quantity:  10,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						uuid.New().String(),
						productID,
						5, // available
						0, // reserved
						0, // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewIncreaseStockUseCase(tt.mockRepo)
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
		})
	}
}
