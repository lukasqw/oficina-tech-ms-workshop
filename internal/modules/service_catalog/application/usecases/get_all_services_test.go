package usecases

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
)

func TestGetAllServicesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockServiceRepository)
		wantErr        error
		validateOutput func(*testing.T, *GetAllServicesOutput)
	}{
		{
			name: "sucesso: retorna lista de múltiplos serviços com todos os campos corretos",
			setupMock: func(m *mockServiceRepository) {
				m.findAllFunc = func(_ context.Context) ([]*service.Service, error) {
					// Create first service
					svc1, err := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
					if err != nil {
						t.Fatalf("failed to create test service 1: %v", err)
					}
					_ = svc1.SetID("550e8400-e29b-41d4-a716-446655440001")

					// Create second service
					svc2, err := service.NewService("Alinhamento", "Alinhamento e balanceamento das rodas", 12000)
					if err != nil {
						t.Fatalf("failed to create test service 2: %v", err)
					}
					_ = svc2.SetID("550e8400-e29b-41d4-a716-446655440002")

					// Create third service
					svc3, err := service.NewService("Revisão Completa", "Revisão completa do veículo com checklist", 50000)
					if err != nil {
						t.Fatalf("failed to create test service 3: %v", err)
					}
					_ = svc3.SetID("550e8400-e29b-41d4-a716-446655440003")

					return []*service.Service{svc1, svc2, svc3}, nil
				}
			},
			wantErr: nil,
			validateOutput: func(t *testing.T, output *GetAllServicesOutput) {
				if output == nil {
					t.Fatal("expected non-nil output")
				}

				if len(output.Services) != 3 {
					t.Fatalf("expected 3 services, got %d", len(output.Services))
				}

				// Validate first service
				svc1 := output.Services[0]
				if svc1.ID != "550e8400-e29b-41d4-a716-446655440001" {
					t.Errorf("Service 1 ID: got %s, want %s", svc1.ID, "550e8400-e29b-41d4-a716-446655440001")
				}
				if svc1.Name != "Troca de Óleo" {
					t.Errorf("Service 1 Name: got %s, want %s", svc1.Name, "Troca de Óleo")
				}
				if svc1.Description != "Troca de óleo do motor com filtro" {
					t.Errorf("Service 1 Description: got %s, want %s", svc1.Description, "Troca de óleo do motor com filtro")
				}
				if svc1.Price != 15000 {
					t.Errorf("Service 1 Price: got %d, want %d", svc1.Price, 15000)
				}
				if svc1.CreatedAt.IsZero() {
					t.Error("Service 1: expected non-zero CreatedAt")
				}
				if svc1.UpdatedAt.IsZero() {
					t.Error("Service 1: expected non-zero UpdatedAt")
				}

				// Validate second service
				svc2 := output.Services[1]
				if svc2.ID != "550e8400-e29b-41d4-a716-446655440002" {
					t.Errorf("Service 2 ID: got %s, want %s", svc2.ID, "550e8400-e29b-41d4-a716-446655440002")
				}
				if svc2.Name != "Alinhamento" {
					t.Errorf("Service 2 Name: got %s, want %s", svc2.Name, "Alinhamento")
				}
				if svc2.Description != "Alinhamento e balanceamento das rodas" {
					t.Errorf("Service 2 Description: got %s, want %s", svc2.Description, "Alinhamento e balanceamento das rodas")
				}
				if svc2.Price != 12000 {
					t.Errorf("Service 2 Price: got %d, want %d", svc2.Price, 12000)
				}
				if svc2.CreatedAt.IsZero() {
					t.Error("Service 2: expected non-zero CreatedAt")
				}
				if svc2.UpdatedAt.IsZero() {
					t.Error("Service 2: expected non-zero UpdatedAt")
				}

				// Validate third service
				svc3 := output.Services[2]
				if svc3.ID != "550e8400-e29b-41d4-a716-446655440003" {
					t.Errorf("Service 3 ID: got %s, want %s", svc3.ID, "550e8400-e29b-41d4-a716-446655440003")
				}
				if svc3.Name != "Revisão Completa" {
					t.Errorf("Service 3 Name: got %s, want %s", svc3.Name, "Revisão Completa")
				}
				if svc3.Description != "Revisão completa do veículo com checklist" {
					t.Errorf("Service 3 Description: got %s, want %s", svc3.Description, "Revisão completa do veículo com checklist")
				}
				if svc3.Price != 50000 {
					t.Errorf("Service 3 Price: got %d, want %d", svc3.Price, 50000)
				}
				if svc3.CreatedAt.IsZero() {
					t.Error("Service 3: expected non-zero CreatedAt")
				}
				if svc3.UpdatedAt.IsZero() {
					t.Error("Service 3: expected non-zero UpdatedAt")
				}
			},
		},
		{
			name: "sucesso: retorna lista vazia quando não há serviços",
			setupMock: func(m *mockServiceRepository) {
				m.findAllFunc = func(_ context.Context) ([]*service.Service, error) {
					return []*service.Service{}, nil
				}
			},
			wantErr: nil,
			validateOutput: func(t *testing.T, output *GetAllServicesOutput) {
				if output == nil {
					t.Fatal("expected non-nil output")
				}

				if output.Services == nil {
					t.Fatal("expected non-nil Services slice")
				}

				if len(output.Services) != 0 {
					t.Errorf("expected empty Services slice, got %d services", len(output.Services))
				}
			},
		},
		{
			name: "erro: FindAll retorna erro",
			setupMock: func(m *mockServiceRepository) {
				m.findAllFunc = func(_ context.Context) ([]*service.Service, error) {
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
			useCase := NewGetAllServicesUseCase(mockRepo)

			// Execute use case
			output, err := useCase.Execute(context.Background())

			// Validate error
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
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
