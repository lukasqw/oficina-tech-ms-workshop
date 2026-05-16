package usecases

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
)

type mockServiceRepository struct {
	saveFunc         func(context.Context, *service.Service) error
	findByIDFunc     func(context.Context, string) (*service.Service, error)
	findAllFunc      func(context.Context) ([]*service.Service, error)
	existsByNameFunc func(context.Context, string) (bool, error)
	deleteFunc       func(context.Context, string) error
}

func (m *mockServiceRepository) Save(ctx context.Context, s *service.Service) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, s)
	}
	return nil
}

func (m *mockServiceRepository) FindByID(ctx context.Context, id string) (*service.Service, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockServiceRepository) FindAll(ctx context.Context) ([]*service.Service, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return []*service.Service{}, nil
}

func (m *mockServiceRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	if m.existsByNameFunc != nil {
		return m.existsByNameFunc(ctx, name)
	}
	return false, nil
}

func (m *mockServiceRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestCreateServiceUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          CreateServiceInput
		setupMock      func(*mockServiceRepository)
		wantErr        error
		validateOutput func(*testing.T, *CreateServiceOutput)
	}{
		{
			name: "sucesso: criação com dados válidos e nome único",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					if name != "Troca de Óleo" {
						t.Errorf("ExistsByName called with wrong name: got %s, want %s", name, "Troca de Óleo")
					}
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					// Simulate repository setting ID after save
					return s.SetID("550e8400-e29b-41d4-a716-446655440000")
				}
			},
			wantErr: nil,
			validateOutput: func(t *testing.T, output *CreateServiceOutput) {
				if output == nil {
					t.Fatal("expected non-nil output")
					return
				}
				if output.ID == "" {
					t.Error("expected non-empty ID")
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
			name: "erro: nome duplicado",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return true, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when name is duplicate")
					return nil
				}
			},
			wantErr: service.ErrDuplicateServiceName,
		},
		{
			name: "erro: nome inválido (muito curto)",
			input: CreateServiceInput{
				Name:        "AB",
				Description: "Descrição válida com mais de 10 caracteres",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidServiceName,
		},
		{
			name: "erro: nome inválido (muito longo)",
			input: CreateServiceInput{
				Name:        "Este é um nome extremamente longo que ultrapassa o limite de 100 caracteres permitidos para o nome do serviço",
				Description: "Descrição válida com mais de 10 caracteres",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidServiceName,
		},
		{
			name: "erro: descrição inválida (muito curta)",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Curta",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidDescription,
		},
		{
			name: "erro: descrição inválida (muito longa)",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Esta é uma descrição extremamente longa que ultrapassa o limite de 500 caracteres permitidos para a descrição do serviço. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidDescription,
		},
		{
			name: "erro: preço inválido (zero)",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       0,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidPrice,
		},
		{
			name: "erro: preço inválido (negativo)",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       -1000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidPrice,
		},
		{
			name: "erro: ExistsByName retorna erro",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, errors.New("database connection error")
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when ExistsByName fails")
					return nil
				}
			},
			wantErr: errors.New("database connection error"),
		},
		{
			name: "erro: Save retorna erro",
			input: CreateServiceInput{
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       15000,
			},
			setupMock: func(m *mockServiceRepository) {
				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					return errors.New("database save error")
				}
			},
			wantErr: errors.New("database save error"),
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
			useCase := NewCreateServiceUseCase(mockRepo)

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
