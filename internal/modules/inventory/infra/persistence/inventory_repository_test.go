//go:build cgo
// +build cgo

package persistence

import (
	"context"
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInventoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&InventoryModel{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func createTestInventory(t *testing.T, productID string, available, reserved, pending int) *inventory.Inventory {
	inv, err := inventory.NewInventory(productID)
	if err != nil {
		t.Fatalf("failed to create test inventory: %v", err)
	}

	// Use Increase to set initial available quantity
	if available > 0 {
		if err := inv.Increase(available); err != nil {
			t.Fatalf("failed to set available quantity: %v", err)
		}
	}

	// Use Reserve to set reserved quantity
	if reserved > 0 {
		if err := inv.Reserve(reserved); err != nil {
			t.Fatalf("failed to set reserved quantity: %v", err)
		}
	}

	// Manually set pending if needed (for reconstruction scenarios)
	// Note: In real scenarios, pending is set through Reserve when insufficient stock

	return inv
}

func TestInventoryRepositoryImpl_Save(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*gorm.DB) *inventory.Inventory
		wantErr   bool
	}{
		{
			name: "should create new inventory successfully",
			setupFunc: func(db *gorm.DB) *inventory.Inventory {
				productID := uuid.New().String()
				return createTestInventory(t, productID, 100, 0, 0)
			},
			wantErr: false,
		},
		{
			name: "should update existing inventory successfully",
			setupFunc: func(db *gorm.DB) *inventory.Inventory {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 50, 10, 5)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save initial inventory: %v", err)
				}

				// Update inventory
				if err := inv.Increase(20); err != nil {
					t.Fatalf("failed to increase stock: %v", err)
				}

				return inv
			},
			wantErr: false,
		},
		{
			name: "should create inventory with zero quantities",
			setupFunc: func(db *gorm.DB) *inventory.Inventory {
				productID := uuid.New().String()
				return createTestInventory(t, productID, 0, 0, 0)
			},
			wantErr: false,
		},
		{
			name: "should create inventory with reserved and pending quantities",
			setupFunc: func(db *gorm.DB) *inventory.Inventory {
				productID := uuid.New().String()
				return createTestInventory(t, productID, 100, 25, 15)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInventoryTestDB(t)
			repo := NewInventoryRepository(db)

			inv := tt.setupFunc(db)
			err := repo.Save(context.Background(), inv)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify inventory was saved
				if inv.ID() == "" {
					t.Error("Save() did not set inventory ID")
				}

				// Verify we can retrieve it
				retrieved, err := repo.FindByID(context.Background(), inv.ID())
				if err != nil {
					t.Errorf("FindByID() after Save() error = %v", err)
					return
				}

				if retrieved.ProductID() != inv.ProductID() {
					t.Errorf("Retrieved inventory ProductID = %v, want %v", retrieved.ProductID(), inv.ProductID())
				}
				if retrieved.AvailableQuantity() != inv.AvailableQuantity() {
					t.Errorf("Retrieved inventory AvailableQuantity = %v, want %v", retrieved.AvailableQuantity(), inv.AvailableQuantity())
				}
			}
		})
	}
}

func TestInventoryRepositoryImpl_FindByID(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should find existing inventory by ID",
			setupFunc: func(db *gorm.DB) string {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 75, 10, 5)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save inventory: %v", err)
				}

				return inv.ID()
			},
			wantErr: false,
		},
		{
			name: "should return error for non-existent inventory",
			setupFunc: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			wantErr:     true,
			expectedErr: inventory.ErrInventoryNotFound,
		},
		{
			name: "should return error for invalid UUID format",
			setupFunc: func(db *gorm.DB) string {
				return "invalid-uuid"
			},
			wantErr:     true,
			expectedErr: inventory.ErrInvalidInventoryID,
		},
		{
			name: "should return error for empty ID",
			setupFunc: func(db *gorm.DB) string {
				return ""
			},
			wantErr:     true,
			expectedErr: inventory.ErrInvalidInventoryID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInventoryTestDB(t)
			repo := NewInventoryRepository(db)

			id := tt.setupFunc(db)
			inv, err := repo.FindByID(context.Background(), id)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("FindByID() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}

			if !tt.wantErr && inv == nil {
				t.Error("FindByID() returned nil inventory")
			}
		})
	}
}

func TestInventoryRepositoryImpl_FindByProductID(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should find existing inventory by product ID",
			setupFunc: func(db *gorm.DB) string {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 50, 5, 3)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save inventory: %v", err)
				}

				return productID
			},
			wantErr: false,
		},
		{
			name: "should return error for non-existent product",
			setupFunc: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			wantErr:     true,
			expectedErr: inventory.ErrInventoryNotFound,
		},
		{
			name: "should return error for invalid product UUID format",
			setupFunc: func(db *gorm.DB) string {
				return "invalid-product-uuid"
			},
			wantErr:     true,
			expectedErr: inventory.ErrInvalidProductID,
		},
		{
			name: "should return error for empty product ID",
			setupFunc: func(db *gorm.DB) string {
				return ""
			},
			wantErr:     true,
			expectedErr: inventory.ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInventoryTestDB(t)
			repo := NewInventoryRepository(db)

			productID := tt.setupFunc(db)
			inv, err := repo.FindByProductID(context.Background(), productID)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindByProductID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("FindByProductID() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}

			if !tt.wantErr {
				if inv == nil {
					t.Error("FindByProductID() returned nil inventory")
					return
				}
				if inv.ProductID() != productID {
					t.Errorf("FindByProductID() ProductID = %v, want %v", inv.ProductID(), productID)
				}
			}
		})
	}
}

func TestInventoryRepositoryImpl_FindAll(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*gorm.DB)
		expectedCount int
		wantErr       bool
	}{
		{
			name: "should return all inventories",
			setupFunc: func(db *gorm.DB) {
				repo := NewInventoryRepository(db)

				inventories := []*inventory.Inventory{
					createTestInventory(t, uuid.New().String(), 100, 10, 5),
					createTestInventory(t, uuid.New().String(), 200, 20, 10),
					createTestInventory(t, uuid.New().String(), 50, 5, 2),
				}

				for _, inv := range inventories {
					if err := repo.Save(context.Background(), inv); err != nil {
						t.Fatalf("failed to save inventory: %v", err)
					}
				}
			},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name: "should return empty list when no inventories exist",
			setupFunc: func(db *gorm.DB) {
				// No inventories created
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "should not return soft-deleted inventories",
			setupFunc: func(db *gorm.DB) {
				repo := NewInventoryRepository(db)

				inv1 := createTestInventory(t, uuid.New().String(), 100, 0, 0)
				inv2 := createTestInventory(t, uuid.New().String(), 200, 0, 0)

				if err := repo.Save(context.Background(), inv1); err != nil {
					t.Fatalf("failed to save inventory 1: %v", err)
				}
				if err := repo.Save(context.Background(), inv2); err != nil {
					t.Fatalf("failed to save inventory 2: %v", err)
				}

				// Delete second inventory
				if err := repo.Delete(context.Background(), inv2.ID()); err != nil {
					t.Fatalf("failed to delete inventory: %v", err)
				}
			},
			expectedCount: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInventoryTestDB(t)
			repo := NewInventoryRepository(db)

			tt.setupFunc(db)
			inventories, err := repo.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(inventories) != tt.expectedCount {
				t.Errorf("FindAll() returned %d inventories, want %d", len(inventories), tt.expectedCount)
			}
		})
	}
}

func TestInventoryRepositoryImpl_Delete(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should soft delete existing inventory",
			setupFunc: func(db *gorm.DB) string {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 100, 0, 0)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save inventory: %v", err)
				}

				return inv.ID()
			},
			wantErr: false,
		},
		{
			name: "should return error for non-existent inventory",
			setupFunc: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			wantErr:     true,
			expectedErr: inventory.ErrInventoryNotFound,
		},
		{
			name: "should return error for invalid UUID format",
			setupFunc: func(db *gorm.DB) string {
				return "invalid-uuid"
			},
			wantErr:     true,
			expectedErr: inventory.ErrInvalidInventoryID,
		},
		{
			name: "should return error when deleting already deleted inventory",
			setupFunc: func(db *gorm.DB) string {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 50, 0, 0)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save inventory: %v", err)
				}

				// Delete once
				if err := repo.Delete(context.Background(), inv.ID()); err != nil {
					t.Fatalf("failed to delete inventory: %v", err)
				}

				return inv.ID()
			},
			wantErr:     true,
			expectedErr: inventory.ErrInventoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInventoryTestDB(t)
			repo := NewInventoryRepository(db)

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
				// Verify inventory is soft deleted
				_, err := repo.FindByID(context.Background(), id)
				if !errors.Is(err, inventory.ErrInventoryNotFound) {
					t.Error("Delete() did not soft delete inventory")
				}
			}
		})
	}
}

func TestInventoryRepositoryImpl_ExistsByProductID(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*gorm.DB) string
		wantExists  bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should return true for existing product inventory",
			setupFunc: func(db *gorm.DB) string {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 100, 0, 0)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save inventory: %v", err)
				}

				return productID
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name: "should return false for non-existent product inventory",
			setupFunc: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "should return error for invalid product UUID format",
			setupFunc: func(db *gorm.DB) string {
				return "invalid-product-uuid"
			},
			wantExists:  false,
			wantErr:     true,
			expectedErr: inventory.ErrInvalidProductID,
		},
		{
			name: "should return error for empty product ID",
			setupFunc: func(db *gorm.DB) string {
				return ""
			},
			wantExists:  false,
			wantErr:     true,
			expectedErr: inventory.ErrInvalidProductID,
		},
		{
			name: "should return false for soft-deleted inventory",
			setupFunc: func(db *gorm.DB) string {
				productID := uuid.New().String()
				inv := createTestInventory(t, productID, 100, 0, 0)

				repo := NewInventoryRepository(db)
				if err := repo.Save(context.Background(), inv); err != nil {
					t.Fatalf("failed to save inventory: %v", err)
				}

				// Delete inventory
				if err := repo.Delete(context.Background(), inv.ID()); err != nil {
					t.Fatalf("failed to delete inventory: %v", err)
				}

				return productID
			},
			wantExists: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInventoryTestDB(t)
			repo := NewInventoryRepository(db)

			productID := tt.setupFunc(db)
			exists, err := repo.ExistsByProductID(context.Background(), productID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExistsByProductID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("ExistsByProductID() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}

			if exists != tt.wantExists {
				t.Errorf("ExistsByProductID() exists = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestInventoryModel_ToDomain(t *testing.T) {
	now := time.Now()
	validInventoryID := uuid.New()
	validProductID := uuid.New()

	tests := []struct {
		name    string
		model   InventoryModel
		wantErr bool
	}{
		{
			name: "should convert valid model to domain",
			model: InventoryModel{
				ID:                validInventoryID,
				ProductID:         validProductID,
				AvailableQuantity: 100,
				ReservedQuantity:  10,
				PendingQuantity:   5,
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			wantErr: false,
		},
		{
			name: "should convert model with zero quantities",
			model: InventoryModel{
				ID:                validInventoryID,
				ProductID:         validProductID,
				AvailableQuantity: 0,
				ReservedQuantity:  0,
				PendingQuantity:   0,
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			wantErr: false,
		},
		{
			name: "should convert model with deleted_at",
			model: InventoryModel{
				ID:                validInventoryID,
				ProductID:         validProductID,
				AvailableQuantity: 50,
				ReservedQuantity:  5,
				PendingQuantity:   2,
				CreatedAt:         now,
				UpdatedAt:         now,
				DeletedAt:         gorm.DeletedAt{Time: now, Valid: true},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := tt.model.ToDomain()

			if (err != nil) != tt.wantErr {
				t.Errorf("ToDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if inv == nil {
					t.Error("ToDomain() returned nil inventory")
					return
				}

				if inv.ID() != tt.model.ID.String() {
					t.Errorf("ToDomain() ID = %v, want %v", inv.ID(), tt.model.ID.String())
				}
				if inv.ProductID() != tt.model.ProductID.String() {
					t.Errorf("ToDomain() ProductID = %v, want %v", inv.ProductID(), tt.model.ProductID.String())
				}
				if inv.AvailableQuantity() != tt.model.AvailableQuantity {
					t.Errorf("ToDomain() AvailableQuantity = %v, want %v", inv.AvailableQuantity(), tt.model.AvailableQuantity)
				}
				if inv.ReservedQuantity() != tt.model.ReservedQuantity {
					t.Errorf("ToDomain() ReservedQuantity = %v, want %v", inv.ReservedQuantity(), tt.model.ReservedQuantity)
				}
				if inv.PendingQuantity() != tt.model.PendingQuantity {
					t.Errorf("ToDomain() PendingQuantity = %v, want %v", inv.PendingQuantity(), tt.model.PendingQuantity)
				}
			}
		})
	}
}

func TestFromDomainInventory(t *testing.T) {
	tests := []struct {
		name      string
		inventory *inventory.Inventory
		wantErr   bool
	}{
		{
			name:      "should convert new inventory without ID",
			inventory: createTestInventory(t, uuid.New().String(), 100, 10, 5),
			wantErr:   false,
		},
		{
			name: "should convert existing inventory with ID",
			inventory: func() *inventory.Inventory {
				inv := createTestInventory(t, uuid.New().String(), 200, 20, 10)
				_ = inv.SetID(uuid.New().String())
				return inv
			}(),
			wantErr: false,
		},
		{
			name:      "should convert inventory with zero quantities",
			inventory: createTestInventory(t, uuid.New().String(), 0, 0, 0),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := FromDomainInventory(tt.inventory)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromDomainInventory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if model == nil {
					t.Error("FromDomainInventory() returned nil model")
					return
				}

				if model.ProductID.String() != tt.inventory.ProductID() {
					t.Errorf("FromDomainInventory() ProductID = %v, want %v", model.ProductID.String(), tt.inventory.ProductID())
				}
				if model.AvailableQuantity != tt.inventory.AvailableQuantity() {
					t.Errorf("FromDomainInventory() AvailableQuantity = %v, want %v", model.AvailableQuantity, tt.inventory.AvailableQuantity())
				}
				if model.ReservedQuantity != tt.inventory.ReservedQuantity() {
					t.Errorf("FromDomainInventory() ReservedQuantity = %v, want %v", model.ReservedQuantity, tt.inventory.ReservedQuantity())
				}
				if model.PendingQuantity != tt.inventory.PendingQuantity() {
					t.Errorf("FromDomainInventory() PendingQuantity = %v, want %v", model.PendingQuantity, tt.inventory.PendingQuantity())
				}
			}
		})
	}
}
