//go:build cgo
// +build cgo

package persistence

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&ProductModel{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func createTestProduct(t *testing.T, name, description string, price int, productType product.ProductType) *product.Product {
	p, err := product.NewProduct(name, description, price, productType)
	if err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}
	return p
}

func TestProductRepositoryImpl_Save(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) *product.Product
		wantErr     bool
		errContains string
	}{
		{
			name: "should create new product successfully",
			setupFunc: func(db *gorm.DB) *product.Product {
				productType, _ := product.NewProductType("CONSUMABLE")
				return createTestProduct(t, "Óleo de Motor", "Óleo sintético 5W30", 5000, productType)
			},
			wantErr: false,
		},
		{
			name: "should update existing product successfully",
			setupFunc: func(db *gorm.DB) *product.Product {
				// Create initial product
				productType, _ := product.NewProductType("PARTS")
				p := createTestProduct(t, "Filtro de Óleo", "Filtro original", 3000, productType)

				repo := NewProductRepository(db)
				if err := repo.Save(context.Background(), p); err != nil {
					t.Fatalf("failed to save initial product: %v", err)
				}

				// Update product
				if err := p.ChangeName("Filtro de Óleo Premium"); err != nil {
					t.Fatalf("failed to update product name: %v", err)
				}

				return p
			},
			wantErr: false,
		},
		{
			name: "should create product with empty description",
			setupFunc: func(db *gorm.DB) *product.Product {
				productType, _ := product.NewProductType("SIMPLE")
				return createTestProduct(t, "Produto Simples", "", 1000, productType)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewProductRepository(db)

			p := tt.setupFunc(db)
			err := repo.Save(context.Background(), p)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify product was saved
				if p.ID() == "" {
					t.Error("Save() did not set product ID")
				}

				// Verify we can retrieve it
				retrieved, err := repo.FindByID(context.Background(), p.ID())
				if err != nil {
					t.Errorf("FindByID() after Save() error = %v", err)
					return
				}

				if retrieved.Name() != p.Name() {
					t.Errorf("Retrieved product name = %v, want %v", retrieved.Name(), p.Name())
				}
			}
		})
	}
}

func TestProductRepositoryImpl_FindByID(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should find existing product by ID",
			setupFunc: func(db *gorm.DB) string {
				productType, _ := product.NewProductType("CONSUMABLE")
				p := createTestProduct(t, "Produto Teste", "Descrição teste", 2500, productType)

				repo := NewProductRepository(db)
				if err := repo.Save(context.Background(), p); err != nil {
					t.Fatalf("failed to save product: %v", err)
				}

				return p.ID()
			},
			wantErr: false,
		},
		{
			name: "should return error for non-existent product",
			setupFunc: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			wantErr:     true,
			expectedErr: product.ErrProductNotFound,
		},
		{
			name: "should return error for invalid UUID format",
			setupFunc: func(db *gorm.DB) string {
				return "invalid-uuid"
			},
			wantErr:     true,
			expectedErr: product.ErrInvalidProductID,
		},
		{
			name: "should return error for empty ID",
			setupFunc: func(db *gorm.DB) string {
				return ""
			},
			wantErr:     true,
			expectedErr: product.ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewProductRepository(db)

			id := tt.setupFunc(db)
			p, err := repo.FindByID(context.Background(), id)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("FindByID() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}

			if !tt.wantErr && p == nil {
				t.Error("FindByID() returned nil product")
			}
		})
	}
}

func TestProductRepositoryImpl_FindAll(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*gorm.DB)
		expectedCount int
		wantErr       bool
	}{
		{
			name: "should return all products",
			setupFunc: func(db *gorm.DB) {
				repo := NewProductRepository(db)
				productType, _ := product.NewProductType("CONSUMABLE")

				products := []*product.Product{
					createTestProduct(t, "Produto 1", "Desc 1", 1000, productType),
					createTestProduct(t, "Produto 2", "Desc 2", 2000, productType),
					createTestProduct(t, "Produto 3", "Desc 3", 3000, productType),
				}

				for _, p := range products {
					if err := repo.Save(context.Background(), p); err != nil {
						t.Fatalf("failed to save product: %v", err)
					}
				}
			},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name: "should return empty list when no products exist",
			setupFunc: func(db *gorm.DB) {
				// No products created
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "should not return soft-deleted products",
			setupFunc: func(db *gorm.DB) {
				repo := NewProductRepository(db)
				productType, _ := product.NewProductType("PARTS")

				p1 := createTestProduct(t, "Produto Ativo", "Ativo", 1000, productType)
				p2 := createTestProduct(t, "Produto Deletado", "Deletado", 2000, productType)

				if err := repo.Save(context.Background(), p1); err != nil {
					t.Fatalf("failed to save product 1: %v", err)
				}
				if err := repo.Save(context.Background(), p2); err != nil {
					t.Fatalf("failed to save product 2: %v", err)
				}

				// Delete second product
				if err := repo.Delete(context.Background(), p2.ID()); err != nil {
					t.Fatalf("failed to delete product: %v", err)
				}
			},
			expectedCount: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewProductRepository(db)

			tt.setupFunc(db)
			products, err := repo.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(products) != tt.expectedCount {
				t.Errorf("FindAll() returned %d products, want %d", len(products), tt.expectedCount)
			}
		})
	}
}

func TestProductRepositoryImpl_Delete(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should soft delete existing product",
			setupFunc: func(db *gorm.DB) string {
				productType, _ := product.NewProductType("CONSUMABLE")
				p := createTestProduct(t, "Produto para Deletar", "Será deletado", 1500, productType)

				repo := NewProductRepository(db)
				if err := repo.Save(context.Background(), p); err != nil {
					t.Fatalf("failed to save product: %v", err)
				}

				return p.ID()
			},
			wantErr: false,
		},
		{
			name: "should return error for non-existent product",
			setupFunc: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			wantErr:     true,
			expectedErr: product.ErrProductNotFound,
		},
		{
			name: "should return error for invalid UUID format",
			setupFunc: func(db *gorm.DB) string {
				return "invalid-uuid"
			},
			wantErr:     true,
			expectedErr: product.ErrInvalidProductID,
		},
		{
			name: "should return error when deleting already deleted product",
			setupFunc: func(db *gorm.DB) string {
				productType, _ := product.NewProductType("SIMPLE")
				p := createTestProduct(t, "Produto Já Deletado", "Deletado", 1000, productType)

				repo := NewProductRepository(db)
				if err := repo.Save(context.Background(), p); err != nil {
					t.Fatalf("failed to save product: %v", err)
				}

				// Delete once
				if err := repo.Delete(context.Background(), p.ID()); err != nil {
					t.Fatalf("failed to delete product: %v", err)
				}

				return p.ID()
			},
			wantErr:     true,
			expectedErr: product.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewProductRepository(db)

			id := tt.setupFunc(db)
			err := repo.Delete(context.Background(), id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Delete() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}

			if !tt.wantErr {
				// Verify product is soft deleted
				_, err := repo.FindByID(context.Background(), id)
				if !errors.Is(err, product.ErrProductNotFound) {
					t.Error("Delete() did not soft delete product")
				}
			}
		})
	}
}

func TestProductModel_ToDomain(t *testing.T) {
	now := time.Now()
	validUUID := uuid.New()

	tests := []struct {
		name        string
		model       ProductModel
		wantErr     bool
		errContains string
	}{
		{
			name: "should convert valid model to domain",
			model: ProductModel{
				ID:          validUUID,
				Name:        "Produto Teste",
				Description: "Descrição teste",
				Price:       5000,
				ProductType: "CONSUMABLE",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: false,
		},
		{
			name: "should convert model with empty description",
			model: ProductModel{
				ID:          validUUID,
				Name:        "Produto Simples",
				Description: "",
				Price:       1000,
				ProductType: "SIMPLE",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: false,
		},
		{
			name: "should return error for invalid product type",
			model: ProductModel{
				ID:          validUUID,
				Name:        "Produto Inválido",
				Description: "Tipo inválido",
				Price:       1000,
				ProductType: "INVALID_TYPE",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "should convert model with deleted_at",
			model: ProductModel{
				ID:          validUUID,
				Name:        "Produto Deletado",
				Description: "Foi deletado",
				Price:       2000,
				ProductType: "PARTS",
				CreatedAt:   now,
				UpdatedAt:   now,
				DeletedAt:   gorm.DeletedAt{Time: now, Valid: true},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := tt.model.ToDomain()

			if (err != nil) != tt.wantErr {
				t.Errorf("ToDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if p == nil {
					t.Error("ToDomain() returned nil product")
					return
				}

				if p.ID() != tt.model.ID.String() {
					t.Errorf("ToDomain() ID = %v, want %v", p.ID(), tt.model.ID.String())
				}
				if p.Name() != tt.model.Name {
					t.Errorf("ToDomain() Name = %v, want %v", p.Name(), tt.model.Name)
				}
				if p.Price() != tt.model.Price {
					t.Errorf("ToDomain() Price = %v, want %v", p.Price(), tt.model.Price)
				}
			}
		})
	}
}

func TestFromDomain(t *testing.T) {
	productType, _ := product.NewProductType("CONSUMABLE")

	tests := []struct {
		name    string
		product *product.Product
		wantErr bool
	}{
		{
			name:    "should convert new product without ID",
			product: createTestProduct(t, "Novo Produto", "Sem ID ainda", 3000, productType),
			wantErr: false,
		},
		{
			name: "should convert existing product with ID",
			product: func() *product.Product {
				p := createTestProduct(t, "Produto Existente", "Com ID", 4000, productType)
				_ = p.SetID(uuid.New().String())
				return p
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := FromDomain(tt.product)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if model == nil {
					t.Error("FromDomain() returned nil model")
					return
				}

				if model.Name != tt.product.Name() {
					t.Errorf("FromDomain() Name = %v, want %v", model.Name, tt.product.Name())
				}
				if model.Price != tt.product.Price() {
					t.Errorf("FromDomain() Price = %v, want %v", model.Price, tt.product.Price())
				}
			}
		})
	}
}
