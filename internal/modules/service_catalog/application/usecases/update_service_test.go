package usecases

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
)

func TestUpdateServiceUseCase_Execute(t *testing.T) {
	// Helper function to create pointer to string
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name      string
		input     UpdateServiceInput
		setupMock func(*mockServiceRepository)
		wantErr   error
		validate  func(*testing.T, *mockServiceRepository)
	}{
		{
			name: "sucesso: atualização de nome apenas com valor válido e único",
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Novo Nome do Serviço"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					if id != "550e8400-e29b-41d4-a716-446655440000" {
						t.Errorf("FindByID called with wrong ID: got %s", id)
					}
					svc, _ := service.NewService("Nome Antigo", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					if name != "Novo Nome do Serviço" {
						t.Errorf("ExistsByName called with wrong name: got %s", name)
					}
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					if s.Name() != "Novo Nome do Serviço" {
						t.Errorf("Save called with wrong name: got %s, want %s", s.Name(), "Novo Nome do Serviço")
					}
					if s.Description() != "Descrição válida com mais de 10 caracteres" {
						t.Errorf("Description should not change: got %s", s.Description())
					}
					if s.Price() != 15000 {
						t.Errorf("Price should not change: got %d", s.Price())
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "sucesso: atualização de descrição apenas",
			input: UpdateServiceInput{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				Description: strPtr("Nova descrição válida com mais de 10 caracteres"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição antiga válida", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					t.Error("ExistsByName should not be called when name is not being updated")
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					if s.Name() != "Nome do Serviço" {
						t.Errorf("Name should not change: got %s", s.Name())
					}
					if s.Description() != "Nova descrição válida com mais de 10 caracteres" {
						t.Errorf("Save called with wrong description: got %s", s.Description())
					}
					if s.Price() != 15000 {
						t.Errorf("Price should not change: got %d", s.Price())
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "sucesso: atualização de preço apenas",
			input: UpdateServiceInput{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Price: intPtr(25000),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					t.Error("ExistsByName should not be called when name is not being updated")
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					if s.Name() != "Nome do Serviço" {
						t.Errorf("Name should not change: got %s", s.Name())
					}
					if s.Description() != "Descrição válida com mais de 10 caracteres" {
						t.Errorf("Description should not change: got %s", s.Description())
					}
					if s.Price() != 25000 {
						t.Errorf("Save called with wrong price: got %d, want %d", s.Price(), 25000)
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "sucesso: atualização de múltiplos campos simultaneamente",
			input: UpdateServiceInput{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				Name:        strPtr("Nome Atualizado"),
				Description: strPtr("Descrição atualizada com mais de 10 caracteres"),
				Price:       intPtr(30000),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome Antigo", "Descrição antiga válida", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					if name != "Nome Atualizado" {
						t.Errorf("ExistsByName called with wrong name: got %s", name)
					}
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					if s.Name() != "Nome Atualizado" {
						t.Errorf("Name not updated: got %s, want %s", s.Name(), "Nome Atualizado")
					}
					if s.Description() != "Descrição atualizada com mais de 10 caracteres" {
						t.Errorf("Description not updated: got %s", s.Description())
					}
					if s.Price() != 30000 {
						t.Errorf("Price not updated: got %d, want %d", s.Price(), 30000)
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "sucesso: atualização de nome para o mesmo valor (não verifica duplicata)",
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Nome do Serviço"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					t.Error("ExistsByName should not be called when name is not changing")
					return false, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					if s.Name() != "Nome do Serviço" {
						t.Errorf("Name should remain: got %s", s.Name())
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "erro: ID inválido",
			input: UpdateServiceInput{
				ID:   "invalid-uuid",
				Name: strPtr("Nome Válido"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					t.Error("FindByID should not be called with invalid UUID")
					return nil, nil
				}
			},
			wantErr: service.ErrInvalidServiceID,
		},
		{
			name: "erro: serviço não encontrado",
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Nome Válido"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, service.ErrServiceNotFound
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when service is not found")
					return nil
				}
			},
			wantErr: service.ErrServiceNotFound,
		},
		{
			name: "erro: nome duplicado ao alterar nome",
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Nome Duplicado"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome Antigo", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.existsByNameFunc = func(_ context.Context, name string) (bool, error) {
					if name == "Nome Duplicado" {
						return true, nil
					}
					return false, nil
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
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("AB"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome Antigo", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

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
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Este é um nome extremamente longo que ultrapassa o limite de 100 caracteres permitidos para o nome do serviço"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome Antigo", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

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
			input: UpdateServiceInput{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				Description: strPtr("Curta"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição antiga válida", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
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
			input: UpdateServiceInput{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				Description: strPtr("Esta é uma descrição extremamente longa que ultrapassa o limite de 500 caracteres permitidos para a descrição do serviço. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição antiga válida", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
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
			input: UpdateServiceInput{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Price: intPtr(0),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
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
			input: UpdateServiceInput{
				ID:    "550e8400-e29b-41d4-a716-446655440000",
				Price: intPtr(-1000),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome do Serviço", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when validation fails")
					return nil
				}
			},
			wantErr: service.ErrInvalidPrice,
		},
		{
			name: "erro: FindByID retorna erro genérico",
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Nome Válido"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					return nil, errors.New("database connection error")
				}

				m.saveFunc = func(_ context.Context, s *service.Service) error {
					t.Error("Save should not be called when FindByID fails")
					return nil
				}
			},
			wantErr: errors.New("database connection error"),
		},
		{
			name: "erro: Save retorna erro",
			input: UpdateServiceInput{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: strPtr("Nome Atualizado"),
			},
			setupMock: func(m *mockServiceRepository) {
				m.findByIDFunc = func(_ context.Context, id string) (*service.Service, error) {
					svc, _ := service.NewService("Nome Antigo", "Descrição válida com mais de 10 caracteres", 15000)
					_ = svc.SetID("550e8400-e29b-41d4-a716-446655440000")
					return svc, nil
				}

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
			useCase := NewUpdateServiceUseCase(mockRepo)

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

			// Additional validations
			if tt.validate != nil {
				tt.validate(t, mockRepo)
			}
		})
	}
}
