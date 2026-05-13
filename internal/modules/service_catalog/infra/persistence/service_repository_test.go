//go:build cgo
// +build cgo

package persistence

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ServiceModelTest is a test-specific model that works with SQLite
type ServiceModelTest struct {
	ID          string `gorm:"type:text;primaryKey"`
	Name        string `gorm:"uniqueIndex;not null;size:100"`
	Description string `gorm:"not null;size:500"`
	Price       int    `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (ServiceModelTest) TableName() string {
	return "services"
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Auto migrate with test model
	if err := db.AutoMigrate(&ServiceModelTest{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// seedService creates a service in the database for testing
func seedService(t *testing.T, db *gorm.DB, name, description string, price int) string {
	t.Helper()
	svc, err := service.NewService(name, description, price)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	repo := NewServiceRepository(db)
	if err := repo.Save(context.Background(), svc); err != nil {
		t.Fatalf("failed to save test service: %v", err)
	}

	return svc.ID()
}

func TestServiceRepository_Save_Create(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		description string
		price       int
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "success - create new service",
			serviceName: "Troca de Óleo",
			description: "Troca de óleo do motor com filtro",
			price:       15000,
			wantErr:     false,
		},
		{
			name:        "success - create service with minimum values",
			serviceName: "ABC",
			description: "1234567890",
			price:       1,
			wantErr:     false,
		},
		{
			name:        "success - create service with maximum name length",
			serviceName: "A" + string(make([]byte, 99)),
			description: "Descrição válida do serviço",
			price:       10000,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewServiceRepository(db)

			svc, err := service.NewService(tt.serviceName, tt.description, tt.price)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}

			err = repo.Save(context.Background(), svc)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && err != tt.expectedErr {
				t.Errorf("Save() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if !tt.wantErr {
				// Verify ID was set
				if svc.ID() == "" {
					t.Error("expected ID to be set after save")
				}

				// Verify service was saved
				found, err := repo.FindByID(context.Background(), svc.ID())
				if err != nil {
					t.Errorf("failed to find saved service: %v", err)
				}
				if found.Name() != tt.serviceName {
					t.Errorf("expected name %s, got %s", tt.serviceName, found.Name())
				}
			}
		})
	}
}

func TestServiceRepository_Save_Update(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T, *gorm.DB) string
		updateFunc  func(*service.Service) error
		wantErr     bool
		expectedErr error
		validate    func(*testing.T, *service.Service)
	}{
		{
			name: "success - update service name",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
			},
			updateFunc: func(s *service.Service) error {
				return s.ChangeName("Troca de Óleo Premium")
			},
			wantErr: false,
			validate: func(t *testing.T, s *service.Service) {
				if s.Name() != "Troca de Óleo Premium" {
					t.Errorf("expected name 'Troca de Óleo Premium', got %s", s.Name())
				}
			},
		},
		{
			name: "success - update service description",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Alinhamento", "Alinhamento básico", 8000)
			},
			updateFunc: func(s *service.Service) error {
				return s.ChangeDescription("Alinhamento e balanceamento completo")
			},
			wantErr: false,
			validate: func(t *testing.T, s *service.Service) {
				if s.Description() != "Alinhamento e balanceamento completo" {
					t.Errorf("expected updated description, got %s", s.Description())
				}
			},
		},
		{
			name: "success - update service price",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Revisão", "Revisão completa do veículo", 50000)
			},
			updateFunc: func(s *service.Service) error {
				return s.ChangePrice(60000)
			},
			wantErr: false,
			validate: func(t *testing.T, s *service.Service) {
				if s.Price() != 60000 {
					t.Errorf("expected price 60000, got %d", s.Price())
				}
			},
		},
		{
			name: "success - update all fields",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Serviço Teste", "Descrição teste", 1000)
			},
			updateFunc: func(s *service.Service) error {
				if err := s.ChangeName("Serviço Atualizado"); err != nil {
					return err
				}
				if err := s.ChangeDescription("Descrição atualizada do serviço"); err != nil {
					return err
				}
				return s.ChangePrice(2000)
			},
			wantErr: false,
			validate: func(t *testing.T, s *service.Service) {
				if s.Name() != "Serviço Atualizado" {
					t.Errorf("expected updated name, got %s", s.Name())
				}
				if s.Description() != "Descrição atualizada do serviço" {
					t.Errorf("expected updated description, got %s", s.Description())
				}
				if s.Price() != 2000 {
					t.Errorf("expected price 2000, got %d", s.Price())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewServiceRepository(db)

			// Setup: create initial service
			id := tt.setupFunc(t, db)

			// Get service
			svc, err := repo.FindByID(context.Background(), id)
			if err != nil {
				t.Fatalf("failed to find service: %v", err)
			}

			// Update service
			if err := tt.updateFunc(svc); err != nil {
				t.Fatalf("failed to update service: %v", err)
			}

			// Save updated service
			err = repo.Save(context.Background(), svc)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && err != tt.expectedErr {
				t.Errorf("Save() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				// Reload from database to verify
				updated, err := repo.FindByID(context.Background(), id)
				if err != nil {
					t.Fatalf("failed to reload service: %v", err)
				}
				tt.validate(t, updated)
			}
		})
	}
}

func TestServiceRepository_Save_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create first service
	svc1, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor", 15000)
	if err := repo.Save(context.Background(), svc1); err != nil {
		t.Fatalf("failed to save first service: %v", err)
	}

	tests := []struct {
		name        string
		serviceName string
		description string
		price       int
		wantErr     bool
	}{
		{
			name:        "error - duplicate name on create",
			serviceName: "Troca de Óleo",
			description: "Outra descrição",
			price:       20000,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := service.NewService(tt.serviceName, tt.description, tt.price)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}

			err = repo.Save(context.Background(), svc)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceRepository_FindByID(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T, *gorm.DB) string
		getID       func(string) string
		wantErr     bool
		expectedErr error
		validate    func(*testing.T, *service.Service)
	}{
		{
			name: "success - find existing service",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
			},
			getID:   func(id string) string { return id },
			wantErr: false,
			validate: func(t *testing.T, s *service.Service) {
				if s.Name() != "Troca de Óleo" {
					t.Errorf("expected name 'Troca de Óleo', got %s", s.Name())
				}
				if s.Price() != 15000 {
					t.Errorf("expected price 15000, got %d", s.Price())
				}
			},
		},
		{
			name: "error - service not found",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return utils.GenerateUUIDv7()
			},
			getID:       func(id string) string { return id },
			wantErr:     true,
			expectedErr: service.ErrServiceNotFound,
		},
		{
			name: "error - invalid UUID format",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return "invalid-uuid"
			},
			getID:       func(id string) string { return id },
			wantErr:     true,
			expectedErr: service.ErrInvalidServiceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewServiceRepository(db)

			id := tt.setupFunc(t, db)
			searchID := tt.getID(id)

			svc, err := repo.FindByID(context.Background(), searchID)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && err != tt.expectedErr {
				t.Errorf("FindByID() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, svc)
			}
		})
	}
}

func TestServiceRepository_FindAll(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T, *gorm.DB)
		wantCount int
		validate  func(*testing.T, []*service.Service)
	}{
		{
			name: "success - find all services",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
				seedService(t, db, "Alinhamento", "Alinhamento e balanceamento", 8000)
				seedService(t, db, "Revisão", "Revisão completa", 50000)
			},
			wantCount: 3,
			validate: func(t *testing.T, services []*service.Service) {
				names := make(map[string]bool)
				for _, s := range services {
					names[s.Name()] = true
				}
				if !names["Troca de Óleo"] || !names["Alinhamento"] || !names["Revisão"] {
					t.Error("expected all services to be returned")
				}
			},
		},
		{
			name: "success - empty list",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				// No services
			},
			wantCount: 0,
		},
		{
			name: "success - find only non-deleted services",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				id1 := seedService(t, db, "Serviço 1", "Descrição 1", 1000)
				seedService(t, db, "Serviço 2", "Descrição 2", 2000)

				// Delete first service
				repo := NewServiceRepository(db)
				_ = repo.Delete(context.Background(), id1)
			},
			wantCount: 1,
			validate: func(t *testing.T, services []*service.Service) {
				if services[0].Name() != "Serviço 2" {
					t.Errorf("expected only non-deleted service, got %s", services[0].Name())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewServiceRepository(db)

			if tt.setupFunc != nil {
				tt.setupFunc(t, db)
			}

			services, err := repo.FindAll(context.Background())

			if err != nil {
				t.Errorf("FindAll() error = %v", err)
				return
			}

			if len(services) != tt.wantCount {
				t.Errorf("FindAll() count = %d, want %d", len(services), tt.wantCount)
				return
			}

			if tt.validate != nil {
				tt.validate(t, services)
			}
		})
	}
}

func TestServiceRepository_ExistsByName(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*testing.T, *gorm.DB)
		searchName string
		wantExists bool
		wantErr    bool
	}{
		{
			name: "success - service exists",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
			},
			searchName: "Troca de Óleo",
			wantExists: true,
			wantErr:    false,
		},
		{
			name: "success - service does not exist",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
			},
			searchName: "Alinhamento",
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "success - case sensitive search",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
			},
			searchName: "troca de óleo",
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "success - deleted service not counted",
			setupFunc: func(t *testing.T, db *gorm.DB) {
				id := seedService(t, db, "Serviço Deletado", "Descrição", 1000)
				repo := NewServiceRepository(db)
				_ = repo.Delete(context.Background(), id)
			},
			searchName: "Serviço Deletado",
			wantExists: false, // GORM filters by deleted_at automatically
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewServiceRepository(db)

			if tt.setupFunc != nil {
				tt.setupFunc(t, db)
			}

			exists, err := repo.ExistsByName(context.Background(), tt.searchName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExistsByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("ExistsByName() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestServiceRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T, *gorm.DB) string
		getID       func(string) string
		wantErr     bool
		expectedErr error
		validate    func(*testing.T, *gorm.DB, string)
	}{
		{
			name: "success - delete existing service",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Troca de Óleo", "Troca de óleo do motor", 15000)
			},
			getID:   func(id string) string { return id },
			wantErr: false,
			validate: func(t *testing.T, db *gorm.DB, id string) {
				repo := NewServiceRepository(db)
				_, err := repo.FindByID(context.Background(), id)
				if err != service.ErrServiceNotFound {
					t.Errorf("expected service to be deleted, but FindByID returned: %v", err)
				}
			},
		},
		{
			name: "success - soft delete preserves data",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return seedService(t, db, "Serviço Teste", "Descrição teste", 1000)
			},
			getID:   func(id string) string { return id },
			wantErr: false,
			validate: func(t *testing.T, db *gorm.DB, id string) {
				// Verify record still exists in database with deleted_at set
				var model ServiceModel
				err := db.Unscoped().First(&model, "id = ?", id).Error
				if err != nil {
					t.Errorf("expected record to exist in database: %v", err)
				}
				if !model.DeletedAt.Valid {
					t.Error("expected deleted_at to be set")
				}
			},
		},
		{
			name: "error - service not found",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return utils.GenerateUUIDv7()
			},
			getID:       func(id string) string { return id },
			wantErr:     true,
			expectedErr: service.ErrServiceNotFound,
		},
		{
			name: "error - invalid UUID format",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				return "invalid-uuid"
			},
			getID:       func(id string) string { return id },
			wantErr:     true,
			expectedErr: service.ErrInvalidServiceID,
		},
		{
			name: "error - delete already deleted service",
			setupFunc: func(t *testing.T, db *gorm.DB) string {
				id := seedService(t, db, "Serviço", "Descrição", 1000)
				repo := NewServiceRepository(db)
				_ = repo.Delete(context.Background(), id)
				return id
			},
			getID:       func(id string) string { return id },
			wantErr:     true,
			expectedErr: service.ErrServiceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewServiceRepository(db)

			id := tt.setupFunc(t, db)
			deleteID := tt.getID(id)

			err := repo.Delete(context.Background(), deleteID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && err != tt.expectedErr {
				t.Errorf("Delete() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, db, id)
			}
		})
	}
}

// TestServiceRepository_Concurrency tests concurrent operations
// Note: SQLite in-memory database has limitations with concurrent access from multiple goroutines
// This test is primarily for documentation and would work better with PostgreSQL
func TestServiceRepository_Concurrency(t *testing.T) {
	t.Skip("Skipping concurrency test with SQLite in-memory database")

	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	id := seedService(t, db, "Serviço Concorrente", "Descrição", 1000)

	t.Run("concurrent reads", func(t *testing.T) {
		done := make(chan error, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := repo.FindByID(context.Background(), id)
				done <- err
			}()
		}

		for i := 0; i < 10; i++ {
			if err := <-done; err != nil {
				t.Errorf("concurrent read failed: %v", err)
			}
		}
	})

	t.Run("concurrent updates", func(t *testing.T) {
		done := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func(index int) {
				svc, err := repo.FindByID(context.Background(), id)
				if err != nil {
					done <- err
					return
				}

				// Update price
				if err := svc.ChangePrice(1000 + index*100); err != nil {
					done <- err
					return
				}
				done <- repo.Save(context.Background(), svc)
			}(i)
		}

		for i := 0; i < 5; i++ {
			if err := <-done; err != nil {
				t.Errorf("concurrent operation failed: %v", err)
			}
		}

		// Verify service still exists and has a valid price
		svc, err := repo.FindByID(context.Background(), id)
		if err != nil {
			t.Errorf("failed to find service after concurrent updates: %v", err)
			return
		}
		if svc.Price() < 1000 || svc.Price() > 1400 {
			t.Errorf("unexpected price after concurrent updates: %d", svc.Price())
		}
	})
}

// TestServiceRepository_Timestamps tests that timestamps are properly managed
func TestServiceRepository_Timestamps(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create service
	svc, _ := service.NewService("Serviço Teste", "Descrição teste", 1000)
	beforeCreate := time.Now().Add(-100 * time.Millisecond)

	if err := repo.Save(context.Background(), svc); err != nil {
		t.Fatalf("failed to save service: %v", err)
	}

	afterCreate := time.Now().Add(100 * time.Millisecond)

	// Verify CreatedAt
	if svc.CreatedAt().Before(beforeCreate) || svc.CreatedAt().After(afterCreate) {
		t.Errorf("CreatedAt not in expected range: %v (expected between %v and %v)",
			svc.CreatedAt(), beforeCreate, afterCreate)
	}

	// Update service
	id := svc.ID()
	time.Sleep(100 * time.Millisecond)
	beforeUpdate := time.Now()

	_ = svc.ChangeName("Nome Atualizado")
	if err := repo.Save(context.Background(), svc); err != nil {
		t.Fatalf("failed to update service: %v", err)
	}

	afterUpdate := time.Now().Add(100 * time.Millisecond)

	// Reload and verify UpdatedAt
	updated, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to reload service: %v", err)
	}

	if updated.UpdatedAt().Before(beforeUpdate) || updated.UpdatedAt().After(afterUpdate) {
		t.Errorf("UpdatedAt not in expected range: %v (expected between %v and %v)",
			updated.UpdatedAt(), beforeUpdate, afterUpdate)
	}

	if !updated.UpdatedAt().After(updated.CreatedAt()) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}
