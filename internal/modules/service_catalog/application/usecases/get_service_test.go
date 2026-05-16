package usecases

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
)

func TestGetServiceByIDUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          GetServiceByIDInput
		setupMock      func(*mockServiceRepository)
		wantErr        error
		validateOutput func(*testing.T, *GetServiceByIDOutput)
	}{
		{
			name: "sucesso: recuperação de serviço existente com todos os campos corretos",
			input: GetServiceByIDInput{
				ID: "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					if id != "550e8400-e29b-41d4-a716-446655440000" {
						t.Errorf("FindByID called with wrong ID: got %s, want %s", id, "550e8400-e29b-41d4-a716-446655440000")
					}

					// Create a service to return
					svc, err := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
					if err != nil {
						t.Fatalf("failed to create test service: %v", err)
					}
					err = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					if err != nil {
						t.Fatalf("failed to set service ID: %v", err)
					}
					return svc, nil
				}
			},
			wantErr: nil,
			validateOutput: func(t *testing.T, output *GetServiceByIDOutput) {
				if output == nil {
					t.Fatal("expected non-nil output")
					return
				}
				if output.ID != "550e8400-e29b-41d4-a716-446655440000" {
					t.Errorf("ID: got %s, want %s", output.ID, "550e8400-e29b-41d4-a716-446655440000")
				}
				if output.Name != "Troca de Óleo" {
					t.Errorf("Name: got %s, want %s", output.Name, "Troca de Óleo")
				}
				if output.Description != "Troca de óleo do motor com filtro" {
					t.Errorf("Description: got %s, want %s", output.Description, "Troca de óleo do motor com filtro")
				}
				if output.Price != 15000 {
					t.Errorf("Price: got %d, want %d", output.Price, 15000)
				}
				if output.CreatedAt.IsZero() {
					t.Error("expected non-zero CreatedAt")
				}
				if output.UpdatedAt.IsZero() {
					t.Error("expected non-zero UpdatedAt")
				}
			},
		},
		{
			name: "erro: ID inválido",
			input: GetServiceByIDInput{
				ID: "invalid-uuid",
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					t.Error("FindByID should not be called when ID is invalid")
					return nil, nil
				}
			},
			wantErr: service.ErrInvalidServiceID,
		},
		{
			name: "erro: ID vazio",
			input: GetServiceByIDInput{
				ID: "",
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					t.Error("FindByID should not be called when ID is empty")
					return nil, nil
				}
			},
			wantErr: service.ErrInvalidServiceID,
		},
		{
			name: "erro: serviço não encontrado",
			input: GetServiceByIDInput{
				ID: "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, service.ErrServiceNotFound
				}
			},
			wantErr: service.ErrServiceNotFound,
		},
		{
			name: "erro: FindByID retorna erro genérico",
			input: GetServiceByIDInput{
				ID: "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, errors.New("database connection error")
				}
			},
			wantErr: errors.New("database connection error"),
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
			useCase := NewGetServiceByIDUseCase(mockRepo)

			// Execute use case
			output, err := useCase.Execute(context.Background(), tt.input)

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

			// Validate output
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}
