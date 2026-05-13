package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"

	"github.com/google/uuid"
)

func TestGetInventoryUseCase_Execute(t *testing.T) {
	validID := uuid.New().String()
	validProductID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name          string
		input         GetInventoryInput
		mockRepo      *mockInventoryRepository
		expectedError error
	}{
		{
			name: "sucesso ao buscar inventário",
			input: GetInventoryInput{
				ID: validID,
			},
			mockRepo: &mockInventoryRepository{
				findByIDFunc: func(_ context.Context, id string) (*inventory.Inventory, error) {
					inv, _ := inventory.ReconstructInventory(
						id,
						validProductID,
						10, // available
						5,  // reserved
						2,  // pending
						now,
						now,
						nil,
					)
					return inv, nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ID inválido",
			input: GetInventoryInput{
				ID: "invalid-uuid",
			},
			mockRepo:      &mockInventoryRepository{},
			expectedError: inventory.ErrInvalidInventoryID,
		},
		{
			name: "erro quando inventário não é encontrado",
			input: GetInventoryInput{
				ID: validID,
			},
			mockRepo: &mockInventoryRepository{
				findByIDFunc: func(_ context.Context, id string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedError: inventory.ErrInventoryNotFound,
		},
		{
			name: "erro ao buscar no repositório",
			input: GetInventoryInput{
				ID: validID,
			},
			mockRepo: &mockInventoryRepository{
				findByIDFunc: func(_ context.Context, id string) (*inventory.Inventory, error) {
					return nil, errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewGetInventoryUseCase(tt.mockRepo)
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

			if output.AvailableQuantity != 10 {
				t.Errorf("esperava AvailableQuantity 10, obteve %d", output.AvailableQuantity)
			}

			if output.ReservedQuantity != 5 {
				t.Errorf("esperava ReservedQuantity 5, obteve %d", output.ReservedQuantity)
			}

			if output.PendingQuantity != 2 {
				t.Errorf("esperava PendingQuantity 2, obteve %d", output.PendingQuantity)
			}
		})
	}
}
