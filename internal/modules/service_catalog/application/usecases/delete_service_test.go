package usecases

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
)

func TestDeleteServiceUseCase_Execute(t *testing.T) {
	validServiceID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name      string
		input     DeleteServiceInput
		setupMock func(*mockServiceRepository)
		wantErr   error
	}{
		{
			name: "sucesso: exclusão de serviço existente",
			input: DeleteServiceInput{
				ID: validServiceID,
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					if id != validServiceID {
						t.Errorf("FindByID called with wrong ID: got %s, want %s", id, validServiceID)
					}

					// Return a valid service
					svc, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
					_ = svc.SetID(validServiceID)
					return svc, nil
				}

				m.deleteFunc = func(_ context.Context, id string) error {
					if id != validServiceID {
						t.Errorf("Delete called with wrong ID: got %s, want %s", id, validServiceID)
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "erro: ID inválido",
			input: DeleteServiceInput{
				ID: "invalid-uuid",
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					t.Error("FindByID should not be called when ID is invalid")
					return nil, nil
				}

				m.deleteFunc = func(_ context.Context, id string) error {
					t.Error("Delete should not be called when ID is invalid")
					return nil
				}
			},
			wantErr: service.ErrInvalidServiceID,
		},
		{
			name: "erro: serviço não encontrado",
			input: DeleteServiceInput{
				ID: validServiceID,
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, service.ErrServiceNotFound
				}

				m.deleteFunc = func(_ context.Context, id string) error {
					t.Error("Delete should not be called when service is not found")
					return nil
				}
			},
			wantErr: service.ErrServiceNotFound,
		},
		{
			name: "erro: FindByID retorna erro genérico",
			input: DeleteServiceInput{
				ID: validServiceID,
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, errors.New("database connection error")
				}

				m.deleteFunc = func(_ context.Context, id string) error {
					t.Error("Delete should not be called when FindByID fails")
					return nil
				}
			},
			wantErr: service.ErrServiceNotFound,
		},
		{
			name: "erro: Delete retorna erro",
			input: DeleteServiceInput{
				ID: validServiceID,
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
					_ = svc.SetID(validServiceID)
					return svc, nil
				}

				m.deleteFunc = func(_ context.Context, id string) error {
					return errors.New("database delete error")
				}
			},
			wantErr: errors.New("database delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := &mockServiceRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			// Create use case with mock
			useCase := NewDeleteServiceUseCase(mockRepo)

			// Execute use case
			err := useCase.Execute(context.Background(), tt.input)

			// Validate error
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			// Validate no error for success cases
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestDeleteServiceUseCase_Execute_CallOrder validates that FindByID is called before Delete
func TestDeleteServiceUseCase_Execute_CallOrder(t *testing.T) {
	validServiceID := "550e8400-e29b-41d4-a716-446655440000"

	var findByIDCalled bool
	var deleteCalled bool
	var callOrder []string

	mockRepo := &mockServiceRepository{
		findByIDFunc: func(_ context.Context, id string) (*service.Service, error) {
			findByIDCalled = true
			callOrder = append(callOrder, "FindByID")

			svc, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
			_ = svc.SetID(validServiceID)
			return svc, nil
		},
		deleteFunc: func(_ context.Context, id string) error {
			deleteCalled = true
			callOrder = append(callOrder, "Delete")
			return nil
		},
	}

	useCase := NewDeleteServiceUseCase(mockRepo)
	err := useCase.Execute(context.Background(), DeleteServiceInput{ID: validServiceID})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findByIDCalled {
		t.Error("FindByID should have been called")
	}

	if !deleteCalled {
		t.Error("Delete should have been called")
	}

	if len(callOrder) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(callOrder))
	}

	if callOrder[0] != "FindByID" {
		t.Errorf("expected first call to be FindByID, got %s", callOrder[0])
	}

	if callOrder[1] != "Delete" {
		t.Errorf("expected second call to be Delete, got %s", callOrder[1])
	}
}

// TestDeleteServiceUseCase_Execute_DeleteNotCalledOnFindByIDFailure validates that Delete is not called when FindByID fails
func TestDeleteServiceUseCase_Execute_DeleteNotCalledOnFindByIDFailure(t *testing.T) {
	validServiceID := "550e8400-e29b-41d4-a716-446655440000"

	var deleteCalled bool

	mockRepo := &mockServiceRepository{
		findByIDFunc: func(_ context.Context, id string) (*service.Service, error) {
			return nil, service.ErrServiceNotFound
		},
		deleteFunc: func(_ context.Context, id string) error {
			deleteCalled = true
			return nil
		},
	}

	useCase := NewDeleteServiceUseCase(mockRepo)
	_ = useCase.Execute(context.Background(), DeleteServiceInput{ID: validServiceID})

	if deleteCalled {
		t.Error("Delete should not have been called when FindByID fails")
	}
}
