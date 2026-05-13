package service

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestNewService tests the NewService constructor with table-driven tests
func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		description string
		price       int
		wantErr     error
	}{
		{
			name:        "sucesso com dados válidos",
			serviceName: "Troca de Óleo",
			description: "Troca de óleo do motor com filtro",
			price:       15000,
			wantErr:     nil,
		},
		{
			name:        "sucesso com nome no limite mínimo (3 caracteres)",
			serviceName: "ABC",
			description: "Descrição válida com mais de 10 caracteres",
			price:       10000,
			wantErr:     nil,
		},
		{
			name:        "sucesso com nome no limite máximo (100 caracteres)",
			serviceName: strings.Repeat("A", 100),
			description: "Descrição válida com mais de 10 caracteres",
			price:       10000,
			wantErr:     nil,
		},
		{
			name:        "sucesso com descrição no limite mínimo (10 caracteres)",
			serviceName: "Serviço Teste",
			description: "1234567890",
			price:       10000,
			wantErr:     nil,
		},
		{
			name:        "sucesso com descrição no limite máximo (500 caracteres)",
			serviceName: "Serviço Teste",
			description: strings.Repeat("A", 500),
			price:       10000,
			wantErr:     nil,
		},
		{
			name:        "sucesso com trimming de espaços no nome",
			serviceName: "  Troca de Óleo  ",
			description: "Troca de óleo do motor com filtro",
			price:       15000,
			wantErr:     nil,
		},
		{
			name:        "sucesso com trimming de espaços na descrição",
			serviceName: "Troca de Óleo",
			description: "  Troca de óleo do motor com filtro  ",
			price:       15000,
			wantErr:     nil,
		},
		{
			name:        "erro: nome inválido (menor que 3 caracteres)",
			serviceName: "AB",
			description: "Descrição válida com mais de 10 caracteres",
			price:       10000,
			wantErr:     ErrInvalidServiceName,
		},
		{
			name:        "erro: nome inválido (maior que 100 caracteres)",
			serviceName: strings.Repeat("A", 101),
			description: "Descrição válida com mais de 10 caracteres",
			price:       10000,
			wantErr:     ErrInvalidServiceName,
		},
		{
			name:        "erro: nome vazio",
			serviceName: "",
			description: "Descrição válida com mais de 10 caracteres",
			price:       10000,
			wantErr:     ErrInvalidServiceName,
		},
		{
			name:        "erro: nome apenas com espaços",
			serviceName: "   ",
			description: "Descrição válida com mais de 10 caracteres",
			price:       10000,
			wantErr:     ErrInvalidServiceName,
		},
		{
			name:        "erro: descrição inválida (menor que 10 caracteres)",
			serviceName: "Serviço Teste",
			description: "Curta",
			price:       10000,
			wantErr:     ErrInvalidDescription,
		},
		{
			name:        "erro: descrição inválida (maior que 500 caracteres)",
			serviceName: "Serviço Teste",
			description: strings.Repeat("A", 501),
			price:       10000,
			wantErr:     ErrInvalidDescription,
		},
		{
			name:        "erro: descrição vazia",
			serviceName: "Serviço Teste",
			description: "",
			price:       10000,
			wantErr:     ErrInvalidDescription,
		},
		{
			name:        "erro: descrição apenas com espaços",
			serviceName: "Serviço Teste",
			description: "     ",
			price:       10000,
			wantErr:     ErrInvalidDescription,
		},
		{
			name:        "erro: preço inválido (zero)",
			serviceName: "Serviço Teste",
			description: "Descrição válida com mais de 10 caracteres",
			price:       0,
			wantErr:     ErrInvalidPrice,
		},
		{
			name:        "erro: preço inválido (negativo)",
			serviceName: "Serviço Teste",
			description: "Descrição válida com mais de 10 caracteres",
			price:       -1000,
			wantErr:     ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCreate := time.Now()
			service, err := NewService(tt.serviceName, tt.description, tt.price)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				}
				if service != nil {
					t.Error("NewService() expected nil service on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewService() unexpected error: %v", err)
			}

			if service == nil {
				t.Fatal("NewService() returned nil service")
			}

			// Validar campos
			expectedName := strings.TrimSpace(tt.serviceName)
			if service.Name() != expectedName {
				t.Errorf("Name() = %v, want %v", service.Name(), expectedName)
			}

			expectedDescription := strings.TrimSpace(tt.description)
			if service.Description() != expectedDescription {
				t.Errorf("Description() = %v, want %v", service.Description(), expectedDescription)
			}

			if service.Price() != tt.price {
				t.Errorf("Price() = %v, want %v", service.Price(), tt.price)
			}

			// Validar timestamps
			if service.CreatedAt().Before(beforeCreate) {
				t.Error("CreatedAt() should be after test start time")
			}

			if service.UpdatedAt().Before(beforeCreate) {
				t.Error("UpdatedAt() should be after test start time")
			}

			if !service.CreatedAt().Equal(service.UpdatedAt()) {
				t.Error("CreatedAt() and UpdatedAt() should be equal for new service")
			}

			// Validar que ID está vazio (será definido após persistência)
			if service.ID() != "" {
				t.Error("ID() should be empty for new service")
			}

			// Validar que deletedAt é nil
			if service.DeletedAt() != nil {
				t.Error("DeletedAt() should be nil for new service")
			}

			if service.IsDeleted() {
				t.Error("IsDeleted() should be false for new service")
			}
		})
	}
}

// TestReconstructService tests the ReconstructService function with table-driven tests
func TestReconstructService(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	now := time.Now()
	deletedTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name        string
		id          string
		serviceName string
		description string
		price       int
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
		wantErr     error
	}{
		{
			name:        "sucesso com UUID válido e sem deletedAt",
			id:          validUUID,
			serviceName: "Troca de Óleo",
			description: "Troca de óleo do motor com filtro",
			price:       15000,
			createdAt:   now.Add(-48 * time.Hour),
			updatedAt:   now.Add(-24 * time.Hour),
			deletedAt:   nil,
			wantErr:     nil,
		},
		{
			name:        "sucesso com UUID válido e com deletedAt",
			id:          validUUID,
			serviceName: "Serviço Deletado",
			description: "Descrição de serviço deletado",
			price:       10000,
			createdAt:   now.Add(-72 * time.Hour),
			updatedAt:   now.Add(-24 * time.Hour),
			deletedAt:   &deletedTime,
			wantErr:     nil,
		},
		{
			name:        "sucesso com outro UUID válido",
			id:          "123e4567-e89b-12d3-a456-426614174000",
			serviceName: "Alinhamento",
			description: "Alinhamento e balanceamento de rodas",
			price:       20000,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     nil,
		},
		{
			name:        "erro: UUID inválido (formato incorreto)",
			id:          "invalid-uuid",
			serviceName: "Serviço Teste",
			description: "Descrição válida",
			price:       10000,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     ErrInvalidServiceID,
		},
		{
			name:        "erro: UUID vazio",
			id:          "",
			serviceName: "Serviço Teste",
			description: "Descrição válida",
			price:       10000,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     ErrInvalidServiceID,
		},
		{
			name:        "erro: UUID com formato parcialmente correto",
			id:          "550e8400-e29b-41d4-a716",
			serviceName: "Serviço Teste",
			description: "Descrição válida",
			price:       10000,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     ErrInvalidServiceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := ReconstructService(
				tt.id,
				tt.serviceName,
				tt.description,
				tt.price,
				tt.createdAt,
				tt.updatedAt,
				tt.deletedAt,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ReconstructService() error = %v, wantErr %v", err, tt.wantErr)
				}
				if service != nil {
					t.Error("ReconstructService() expected nil service on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ReconstructService() unexpected error: %v", err)
			}

			if service == nil {
				t.Fatal("ReconstructService() returned nil service")
			}

			// Validar todos os campos
			if service.ID() != tt.id {
				t.Errorf("ID() = %v, want %v", service.ID(), tt.id)
			}

			if service.Name() != tt.serviceName {
				t.Errorf("Name() = %v, want %v", service.Name(), tt.serviceName)
			}

			if service.Description() != tt.description {
				t.Errorf("Description() = %v, want %v", service.Description(), tt.description)
			}

			if service.Price() != tt.price {
				t.Errorf("Price() = %v, want %v", service.Price(), tt.price)
			}

			if !service.CreatedAt().Equal(tt.createdAt) {
				t.Errorf("CreatedAt() = %v, want %v", service.CreatedAt(), tt.createdAt)
			}

			if !service.UpdatedAt().Equal(tt.updatedAt) {
				t.Errorf("UpdatedAt() = %v, want %v", service.UpdatedAt(), tt.updatedAt)
			}

			// Validar deletedAt
			if tt.deletedAt == nil {
				if service.DeletedAt() != nil {
					t.Errorf("DeletedAt() = %v, want nil", service.DeletedAt())
				}
				if service.IsDeleted() {
					t.Error("IsDeleted() should be false when deletedAt is nil")
				}
			} else {
				if service.DeletedAt() == nil {
					t.Error("DeletedAt() should not be nil")
				} else if !service.DeletedAt().Equal(*tt.deletedAt) {
					t.Errorf("DeletedAt() = %v, want %v", service.DeletedAt(), tt.deletedAt)
				}
				if !service.IsDeleted() {
					t.Error("IsDeleted() should be true when deletedAt is set")
				}
			}
		})
	}
}

// TestService_ChangeName tests the ChangeName method with table-driven tests
func TestService_ChangeName(t *testing.T) {
	tests := []struct {
		name     string
		newName  string
		wantErr  error
		wantName string
	}{
		{
			name:     "sucesso: atualização com nome válido",
			newName:  "Novo Nome do Serviço",
			wantErr:  nil,
			wantName: "Novo Nome do Serviço",
		},
		{
			name:     "sucesso: nome no limite mínimo (3 caracteres)",
			newName:  "ABC",
			wantErr:  nil,
			wantName: "ABC",
		},
		{
			name:     "sucesso: nome no limite máximo (100 caracteres)",
			newName:  strings.Repeat("A", 100),
			wantErr:  nil,
			wantName: strings.Repeat("A", 100),
		},
		{
			name:     "sucesso: trimming de espaços",
			newName:  "  Nome com Espaços  ",
			wantErr:  nil,
			wantName: "Nome com Espaços",
		},
		{
			name:     "erro: nome inválido (menor que 3 caracteres)",
			newName:  "AB",
			wantErr:  ErrInvalidServiceName,
			wantName: "Nome Original",
		},
		{
			name:     "erro: nome inválido (maior que 100 caracteres)",
			newName:  strings.Repeat("A", 101),
			wantErr:  ErrInvalidServiceName,
			wantName: "Nome Original",
		},
		{
			name:     "erro: nome vazio",
			newName:  "",
			wantErr:  ErrInvalidServiceName,
			wantName: "Nome Original",
		},
		{
			name:     "erro: nome apenas com espaços",
			newName:  "     ",
			wantErr:  ErrInvalidServiceName,
			wantName: "Nome Original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar serviço inicial
			service, err := NewService("Nome Original", "Descrição original do serviço", 10000)
			if err != nil {
				t.Fatalf("Failed to create service: %v", err)
			}

			originalUpdatedAt := service.UpdatedAt()
			time.Sleep(2 * time.Millisecond) // Garantir que updatedAt será diferente

			// Executar ChangeName
			err = service.ChangeName(tt.newName)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ChangeName() error = %v, wantErr %v", err, tt.wantErr)
				}
				// Validar que o estado não foi modificado
				if service.Name() != "Nome Original" {
					t.Errorf("Name should not change on error, got %v", service.Name())
				}
				if !service.UpdatedAt().Equal(originalUpdatedAt) {
					t.Error("UpdatedAt should not change on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ChangeName() unexpected error: %v", err)
			}

			// Validar que o nome foi atualizado
			if service.Name() != tt.wantName {
				t.Errorf("Name() = %v, want %v", service.Name(), tt.wantName)
			}

			// Validar que updatedAt foi atualizado
			if !service.UpdatedAt().After(originalUpdatedAt) {
				t.Error("UpdatedAt should be updated after ChangeName")
			}

			// Validar que outros campos não foram modificados
			if service.Description() != "Descrição original do serviço" {
				t.Error("Description should not change")
			}
			if service.Price() != 10000 {
				t.Error("Price should not change")
			}
		})
	}
}

// TestService_ChangeDescription tests the ChangeDescription method with table-driven tests
func TestService_ChangeDescription(t *testing.T) {
	tests := []struct {
		name            string
		newDescription  string
		wantErr         error
		wantDescription string
	}{
		{
			name:            "sucesso: atualização com descrição válida",
			newDescription:  "Nova descrição do serviço com detalhes",
			wantErr:         nil,
			wantDescription: "Nova descrição do serviço com detalhes",
		},
		{
			name:            "sucesso: descrição no limite mínimo (10 caracteres)",
			newDescription:  "1234567890",
			wantErr:         nil,
			wantDescription: "1234567890",
		},
		{
			name:            "sucesso: descrição no limite máximo (500 caracteres)",
			newDescription:  strings.Repeat("A", 500),
			wantErr:         nil,
			wantDescription: strings.Repeat("A", 500),
		},
		{
			name:            "sucesso: trimming de espaços",
			newDescription:  "  Descrição com espaços nas pontas  ",
			wantErr:         nil,
			wantDescription: "Descrição com espaços nas pontas",
		},
		{
			name:            "erro: descrição inválida (menor que 10 caracteres)",
			newDescription:  "Curta",
			wantErr:         ErrInvalidDescription,
			wantDescription: "Descrição original do serviço",
		},
		{
			name:            "erro: descrição inválida (maior que 500 caracteres)",
			newDescription:  strings.Repeat("A", 501),
			wantErr:         ErrInvalidDescription,
			wantDescription: "Descrição original do serviço",
		},
		{
			name:            "erro: descrição vazia",
			newDescription:  "",
			wantErr:         ErrInvalidDescription,
			wantDescription: "Descrição original do serviço",
		},
		{
			name:            "erro: descrição apenas com espaços",
			newDescription:  "     ",
			wantErr:         ErrInvalidDescription,
			wantDescription: "Descrição original do serviço",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar serviço inicial
			service, err := NewService("Nome do Serviço", "Descrição original do serviço", 10000)
			if err != nil {
				t.Fatalf("Failed to create service: %v", err)
			}

			originalUpdatedAt := service.UpdatedAt()
			time.Sleep(2 * time.Millisecond) // Garantir que updatedAt será diferente

			// Executar ChangeDescription
			err = service.ChangeDescription(tt.newDescription)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ChangeDescription() error = %v, wantErr %v", err, tt.wantErr)
				}
				// Validar que o estado não foi modificado
				if service.Description() != "Descrição original do serviço" {
					t.Errorf("Description should not change on error, got %v", service.Description())
				}
				if !service.UpdatedAt().Equal(originalUpdatedAt) {
					t.Error("UpdatedAt should not change on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ChangeDescription() unexpected error: %v", err)
			}

			// Validar que a descrição foi atualizada
			if service.Description() != tt.wantDescription {
				t.Errorf("Description() = %v, want %v", service.Description(), tt.wantDescription)
			}

			// Validar que updatedAt foi atualizado
			if !service.UpdatedAt().After(originalUpdatedAt) {
				t.Error("UpdatedAt should be updated after ChangeDescription")
			}

			// Validar que outros campos não foram modificados
			if service.Name() != "Nome do Serviço" {
				t.Error("Name should not change")
			}
			if service.Price() != 10000 {
				t.Error("Price should not change")
			}
		})
	}
}

// TestService_ChangePrice tests the ChangePrice method with table-driven tests
func TestService_ChangePrice(t *testing.T) {
	tests := []struct {
		name      string
		newPrice  int
		wantErr   error
		wantPrice int
	}{
		{
			name:      "sucesso: atualização com preço válido",
			newPrice:  25000,
			wantErr:   nil,
			wantPrice: 25000,
		},
		{
			name:      "sucesso: preço mínimo válido (1)",
			newPrice:  1,
			wantErr:   nil,
			wantPrice: 1,
		},
		{
			name:      "sucesso: preço alto",
			newPrice:  999999,
			wantErr:   nil,
			wantPrice: 999999,
		},
		{
			name:      "erro: preço inválido (zero)",
			newPrice:  0,
			wantErr:   ErrInvalidPrice,
			wantPrice: 10000,
		},
		{
			name:      "erro: preço inválido (negativo)",
			newPrice:  -5000,
			wantErr:   ErrInvalidPrice,
			wantPrice: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar serviço inicial
			service, err := NewService("Nome do Serviço", "Descrição original do serviço", 10000)
			if err != nil {
				t.Fatalf("Failed to create service: %v", err)
			}

			originalUpdatedAt := service.UpdatedAt()
			time.Sleep(2 * time.Millisecond) // Garantir que updatedAt será diferente

			// Executar ChangePrice
			err = service.ChangePrice(tt.newPrice)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ChangePrice() error = %v, wantErr %v", err, tt.wantErr)
				}
				// Validar que o estado não foi modificado
				if service.Price() != 10000 {
					t.Errorf("Price should not change on error, got %v", service.Price())
				}
				if !service.UpdatedAt().Equal(originalUpdatedAt) {
					t.Error("UpdatedAt should not change on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ChangePrice() unexpected error: %v", err)
			}

			// Validar que o preço foi atualizado
			if service.Price() != tt.wantPrice {
				t.Errorf("Price() = %v, want %v", service.Price(), tt.wantPrice)
			}

			// Validar que updatedAt foi atualizado
			if !service.UpdatedAt().After(originalUpdatedAt) {
				t.Error("UpdatedAt should be updated after ChangePrice")
			}

			// Validar que outros campos não foram modificados
			if service.Name() != "Nome do Serviço" {
				t.Error("Name should not change")
			}
			if service.Description() != "Descrição original do serviço" {
				t.Error("Description should not change")
			}
		})
	}
}

// TestService_MarkAsDeleted tests the MarkAsDeleted method
func TestService_MarkAsDeleted(t *testing.T) {
	// Criar serviço inicial
	service, err := NewService("Nome do Serviço", "Descrição do serviço para teste", 10000)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Validar estado inicial
	if service.DeletedAt() != nil {
		t.Error("DeletedAt should be nil initially")
	}
	if service.IsDeleted() {
		t.Error("IsDeleted() should be false initially")
	}

	originalUpdatedAt := service.UpdatedAt()
	time.Sleep(2 * time.Millisecond) // Garantir que updatedAt será diferente

	beforeDelete := time.Now()

	// Executar MarkAsDeleted
	service.MarkAsDeleted()

	// Validar que deletedAt foi definido
	if service.DeletedAt() == nil {
		t.Fatal("DeletedAt should not be nil after MarkAsDeleted")
	}

	// Validar que deletedAt está no momento correto
	if service.DeletedAt().Before(beforeDelete) {
		t.Error("DeletedAt should be after the deletion call")
	}

	// Validar que IsDeleted retorna true
	if !service.IsDeleted() {
		t.Error("IsDeleted() should be true after MarkAsDeleted")
	}

	// Validar que updatedAt foi atualizado
	if !service.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after MarkAsDeleted")
	}

	// Validar que updatedAt e deletedAt são aproximadamente iguais
	timeDiff := service.UpdatedAt().Sub(*service.DeletedAt())
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Error("UpdatedAt and DeletedAt should be set at approximately the same time")
	}

	// Validar que outros campos não foram modificados
	if service.Name() != "Nome do Serviço" {
		t.Error("Name should not change")
	}
	if service.Description() != "Descrição do serviço para teste" {
		t.Error("Description should not change")
	}
	if service.Price() != 10000 {
		t.Error("Price should not change")
	}
}

// TestService_SetID tests the SetID method with table-driven tests
func TestService_SetID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "sucesso: UUID válido",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: nil,
		},
		{
			name:    "sucesso: outro UUID válido",
			id:      "123e4567-e89b-12d3-a456-426614174000",
			wantErr: nil,
		},
		{
			name:    "erro: UUID inválido (formato incorreto)",
			id:      "invalid-uuid",
			wantErr: ErrInvalidServiceID,
		},
		{
			name:    "erro: UUID vazio",
			id:      "",
			wantErr: ErrInvalidServiceID,
		},
		{
			name:    "erro: UUID parcialmente correto",
			id:      "550e8400-e29b-41d4-a716",
			wantErr: ErrInvalidServiceID,
		},
		{
			name:    "erro: string aleatória",
			id:      "not-a-uuid-at-all",
			wantErr: ErrInvalidServiceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar serviço inicial
			service, err := NewService("Nome do Serviço", "Descrição do serviço para teste", 10000)
			if err != nil {
				t.Fatalf("Failed to create service: %v", err)
			}

			// Validar que ID está vazio inicialmente
			if service.ID() != "" {
				t.Error("ID should be empty initially")
			}

			// Executar SetID
			err = service.SetID(tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SetID() error = %v, wantErr %v", err, tt.wantErr)
				}
				// Validar que ID não foi modificado em caso de erro
				if service.ID() != "" {
					t.Errorf("ID should remain empty on error, got %v", service.ID())
				}
				return
			}

			if err != nil {
				t.Fatalf("SetID() unexpected error: %v", err)
			}

			// Validar que ID foi definido corretamente
			if service.ID() != tt.id {
				t.Errorf("ID() = %v, want %v", service.ID(), tt.id)
			}

			// Validar que outros campos não foram modificados
			if service.Name() != "Nome do Serviço" {
				t.Error("Name should not change")
			}
			if service.Description() != "Descrição do serviço para teste" {
				t.Error("Description should not change")
			}
			if service.Price() != 10000 {
				t.Error("Price should not change")
			}
		})
	}
}
