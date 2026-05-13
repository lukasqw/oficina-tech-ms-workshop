package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"

	"github.com/google/uuid"
)

func TestDeleteInventoryUseCase_Execute(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name          string
		input         DeleteInventoryInput
		mockRepo      *mockInventoryRepository
		expectedError error
	}{
		{
			name: "sucesso ao deletar inventário",
			input: DeleteInventoryInput{
				ID: validID,
			},
			mockRepo: &mockInventoryRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			expectedError: nil,
		},
		{
			name: "erro ao fornecer ID inválido",
			input: DeleteInventoryInput{
				ID: "invalid-uuid",
			},
			mockRepo:      &mockInventoryRepository{},
			expectedError: inventory.ErrInvalidInventoryID,
		},
		{
			name: "erro quando inventário não é encontrado",
			input: DeleteInventoryInput{
				ID: validID,
			},
			mockRepo: &mockInventoryRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return inventory.ErrInventoryNotFound
				},
			},
			expectedError: inventory.ErrInventoryNotFound,
		},
		{
			name: "erro ao deletar no repositório",
			input: DeleteInventoryInput{
				ID: validID,
			},
			mockRepo: &mockInventoryRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return errors.New("database error")
				},
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewDeleteInventoryUseCase(tt.mockRepo)
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
