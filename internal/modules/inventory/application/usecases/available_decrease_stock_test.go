package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"

	"github.com/google/uuid"
)

// Mock do repositório de inventory
type mockInventoryRepository struct {
	findByProductIDFunc   func(context.Context, string) (*inventory.Inventory, error)
	saveFunc              func(context.Context, *inventory.Inventory) error
	findByIDFunc          func(context.Context, string) (*inventory.Inventory, error)
	findAllFunc           func(context.Context) ([]*inventory.Inventory, error)
	deleteFunc            func(context.Context, string) error
	existsByProductIDFunc func(context.Context, string) (bool, error)
}

func (m *mockInventoryRepository) FindByProductID(ctx context.Context, productID string) (*inventory.Inventory, error) {
	if m.findByProductIDFunc != nil {
		return m.findByProductIDFunc(ctx, productID)
	}
	return nil, nil
}

func (m *mockInventoryRepository) Save(ctx context.Context, inv *inventory.Inventory) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, inv)
	}
	return nil
}

func (m *mockInventoryRepository) FindByID(ctx context.Context, id string) (*inventory.Inventory, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockInventoryRepository) FindAll(ctx context.Context) ([]*inventory.Inventory, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockInventoryRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockInventoryRepository) ExistsByProductID(ctx context.Context, productID string) (bool, error) {
	if m.existsByProductIDFunc != nil {
		return m.existsByProductIDFunc(ctx, productID)
	}
	return false, nil
}

func TestAvailableDecreaseStockUseCase_Execute(t *testing.T) {
	validProductID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name          string
		input         AvailableDecreaseStockInput
		mockRepo      *mockInventoryRepository
		expectedError error
	}{
		{
			name: "sucesso ao diminuir estoque disponível",
			input: AvailableDecreaseStockInput{
				ProductID: validProductID,
				Quantity:  5,
			},
			mockRepo: &mockInventoryRepository{
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
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ProductID inválido",
			input: AvailableDecreaseStockInput{
				ProductID: "invalid-uuid",
				Quantity:  5,
			},
			mockRepo:      &mockInventoryRepository{},
			expectedError: inventory.ErrInvalidProductID,
		},
		{
			name: "erro quando estoque não é encontrado",
			input: AvailableDecreaseStockInput{
				ProductID: validProductID,
				Quantity:  5,
			},
			mockRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedError: inventory.ErrInventoryNotFound,
		},
		{
			name: "erro quando quantidade disponível é insuficiente",
			input: AvailableDecreaseStockInput{
				ProductID: validProductID,
				Quantity:  15,
			},
			mockRepo: &mockInventoryRepository{
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
			},
			expectedError: inventory.ErrInsufficientAvailable,
		},
		{
			name: "erro ao salvar no repositório",
			input: AvailableDecreaseStockInput{
				ProductID: validProductID,
				Quantity:  5,
			},
			mockRepo: &mockInventoryRepository{
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
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "erro ao fornecer quantidade inválida (zero)",
			input: AvailableDecreaseStockInput{
				ProductID: validProductID,
				Quantity:  0,
			},
			mockRepo: &mockInventoryRepository{
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
			},
			expectedError: inventory.ErrInvalidQuantity,
		},
		{
			name: "erro ao fornecer quantidade negativa",
			input: AvailableDecreaseStockInput{
				ProductID: validProductID,
				Quantity:  -5,
			},
			mockRepo: &mockInventoryRepository{
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
			},
			expectedError: inventory.ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewAvailableDecreaseStockUseCase(tt.mockRepo)
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

			if output.AvailableQuantity != 5 {
				t.Errorf("esperava AvailableQuantity 5, obteve %d", output.AvailableQuantity)
			}
		})
	}
}
