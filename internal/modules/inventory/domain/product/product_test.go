package product

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"testing"
	"time"
)

func TestNewProduct(t *testing.T) {
	validProductType, _ := NewProductType(ProductTypeConsumable)

	tests := []struct {
		name        string
		productName string
		description string
		price       int
		productType ProductType
		wantErr     error
	}{
		{
			name:        "valid product with all fields",
			productName: "Óleo Motor 5W30",
			description: "Óleo sintético para motor",
			price:       5000,
			productType: validProductType,
			wantErr:     nil,
		},
		{
			name:        "valid product with minimum name length",
			productName: "AB",
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     nil,
		},
		{
			name:        "valid product with maximum name length",
			productName: string(make([]byte, 200)),
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     nil,
		},
		{
			name:        "valid product with empty description",
			productName: "Product",
			description: "",
			price:       100,
			productType: validProductType,
			wantErr:     nil,
		},
		{
			name:        "valid product with name containing spaces",
			productName: "  Product Name  ",
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     nil,
		},
		{
			name:        "invalid product - name too short",
			productName: "A",
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     ErrInvalidProductName,
		},
		{
			name:        "invalid product - name too long",
			productName: string(make([]byte, 201)),
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     ErrInvalidProductName,
		},
		{
			name:        "invalid product - empty name",
			productName: "",
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     ErrInvalidProductName,
		},
		{
			name:        "invalid product - name only spaces",
			productName: "   ",
			description: "Test",
			price:       100,
			productType: validProductType,
			wantErr:     ErrInvalidProductName,
		},
		{
			name:        "invalid product - description too long",
			productName: "Product",
			description: string(make([]byte, 1001)),
			price:       100,
			productType: validProductType,
			wantErr:     ErrInvalidDescription,
		},
		{
			name:        "invalid product - zero price",
			productName: "Product",
			description: "Test",
			price:       0,
			productType: validProductType,
			wantErr:     ErrInvalidPrice,
		},
		{
			name:        "invalid product - negative price",
			productName: "Product",
			description: "Test",
			price:       -100,
			productType: validProductType,
			wantErr:     ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(tt.productName, tt.description, tt.price, tt.productType)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewProduct() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("NewProduct() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProduct() unexpected error = %v", err)
				return
			}

			if product == nil {
				t.Error("NewProduct() returned nil product")
				return
			}

			// Verify fields
			if product.Name() != tt.productName && product.Name() != string(make([]byte, 200)) {
				// Handle trimmed name
				trimmedName := tt.productName
				if tt.productName == "  Product Name  " {
					trimmedName = "Product Name"
				}
				if product.Name() != trimmedName {
					t.Errorf("NewProduct() name = %v, want %v", product.Name(), trimmedName)
				}
			}

			if product.Description() != tt.description {
				t.Errorf("NewProduct() description = %v, want %v", product.Description(), tt.description)
			}

			if product.Price() != tt.price {
				t.Errorf("NewProduct() price = %v, want %v", product.Price(), tt.price)
			}

			if product.ProductType().Value() != tt.productType.Value() {
				t.Errorf("NewProduct() productType = %v, want %v", product.ProductType(), tt.productType)
			}

			if product.ID() != "" {
				t.Errorf("NewProduct() id should be empty, got %v", product.ID())
			}

			if product.IsDeleted() {
				t.Error("NewProduct() product should not be deleted")
			}
		})
	}
}

func TestReconstructProduct(t *testing.T) {
	validUUID := utils.GenerateUUIDv7()
	validProductType, _ := NewProductType(ProductTypeSimple)
	now := time.Now()
	deletedTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name        string
		id          string
		productName string
		description string
		price       int
		productType ProductType
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
		wantErr     error
	}{
		{
			name:        "valid reconstruction",
			id:          validUUID,
			productName: "Product",
			description: "Description",
			price:       1000,
			productType: validProductType,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     nil,
		},
		{
			name:        "valid reconstruction with deleted product",
			id:          validUUID,
			productName: "Product",
			description: "Description",
			price:       1000,
			productType: validProductType,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   &deletedTime,
			wantErr:     nil,
		},
		{
			name:        "invalid reconstruction - invalid UUID",
			id:          "invalid-uuid",
			productName: "Product",
			description: "Description",
			price:       1000,
			productType: validProductType,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     ErrInvalidProductID,
		},
		{
			name:        "invalid reconstruction - empty UUID",
			id:          "",
			productName: "Product",
			description: "Description",
			price:       1000,
			productType: validProductType,
			createdAt:   now,
			updatedAt:   now,
			deletedAt:   nil,
			wantErr:     ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := ReconstructProduct(
				tt.id,
				tt.productName,
				tt.description,
				tt.price,
				tt.productType,
				tt.createdAt,
				tt.updatedAt,
				tt.deletedAt,
			)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ReconstructProduct() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("ReconstructProduct() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReconstructProduct() unexpected error = %v", err)
				return
			}

			if product == nil {
				t.Error("ReconstructProduct() returned nil product")
				return
			}

			// Verify all fields
			if product.ID() != tt.id {
				t.Errorf("ReconstructProduct() id = %v, want %v", product.ID(), tt.id)
			}

			if product.Name() != tt.productName {
				t.Errorf("ReconstructProduct() name = %v, want %v", product.Name(), tt.productName)
			}

			if product.Description() != tt.description {
				t.Errorf("ReconstructProduct() description = %v, want %v", product.Description(), tt.description)
			}

			if product.Price() != tt.price {
				t.Errorf("ReconstructProduct() price = %v, want %v", product.Price(), tt.price)
			}

			if product.IsDeleted() != (tt.deletedAt != nil) {
				t.Errorf("ReconstructProduct() isDeleted = %v, want %v", product.IsDeleted(), tt.deletedAt != nil)
			}
		})
	}
}

func TestProduct_SetID(t *testing.T) {
	validProductType, _ := NewProductType(ProductTypeConsumable)
	product, _ := NewProduct("Product", "Description", 1000, validProductType)
	validUUID := utils.GenerateUUIDv7()

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid UUID",
			id:      validUUID,
			wantErr: nil,
		},
		{
			name:    "invalid UUID",
			id:      "invalid-uuid",
			wantErr: ErrInvalidProductID,
		},
		{
			name:    "empty UUID",
			id:      "",
			wantErr: ErrInvalidProductID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := product.SetID(tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("SetID() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("SetID() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("SetID() unexpected error = %v", err)
				return
			}

			if product.ID() != tt.id {
				t.Errorf("SetID() id = %v, want %v", product.ID(), tt.id)
			}
		})
	}
}

func TestProduct_ChangeName(t *testing.T) {
	validProductType, _ := NewProductType(ProductTypeConsumable)

	tests := []struct {
		name         string
		newName      string
		wantErr      error
		expectedName string
	}{
		{
			name:         "valid name change",
			newName:      "New Product Name",
			wantErr:      nil,
			expectedName: "New Product Name",
		},
		{
			name:         "valid name with minimum length",
			newName:      "AB",
			wantErr:      nil,
			expectedName: "AB",
		},
		{
			name:         "valid name with maximum length",
			newName:      string(make([]byte, 200)),
			wantErr:      nil,
			expectedName: string(make([]byte, 200)),
		},
		{
			name:         "valid name with spaces trimmed",
			newName:      "  Trimmed Name  ",
			wantErr:      nil,
			expectedName: "Trimmed Name",
		},
		{
			name:    "invalid name - too short",
			newName: "A",
			wantErr: ErrInvalidProductName,
		},
		{
			name:    "invalid name - too long",
			newName: string(make([]byte, 201)),
			wantErr: ErrInvalidProductName,
		},
		{
			name:    "invalid name - empty",
			newName: "",
			wantErr: ErrInvalidProductName,
		},
		{
			name:    "invalid name - only spaces",
			newName: "   ",
			wantErr: ErrInvalidProductName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("Original Name", "Description", 1000, validProductType)
			oldUpdatedAt := product.UpdatedAt()
			time.Sleep(1 * time.Millisecond) // Ensure time difference

			err := product.ChangeName(tt.newName)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ChangeName() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("ChangeName() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ChangeName() unexpected error = %v", err)
				return
			}

			if product.Name() != tt.expectedName {
				t.Errorf("ChangeName() name = %v, want %v", product.Name(), tt.expectedName)
			}

			if !product.UpdatedAt().After(oldUpdatedAt) {
				t.Error("ChangeName() should update updatedAt timestamp")
			}
		})
	}
}

func TestProduct_ChangeDescription(t *testing.T) {
	validProductType, _ := NewProductType(ProductTypeConsumable)

	tests := []struct {
		name           string
		newDescription string
		wantErr        error
	}{
		{
			name:           "valid description change",
			newDescription: "New description",
			wantErr:        nil,
		},
		{
			name:           "valid empty description",
			newDescription: "",
			wantErr:        nil,
		},
		{
			name:           "valid maximum length description",
			newDescription: string(make([]byte, 1000)),
			wantErr:        nil,
		},
		{
			name:           "invalid description - too long",
			newDescription: string(make([]byte, 1001)),
			wantErr:        ErrInvalidDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("Product", "Original Description", 1000, validProductType)
			oldUpdatedAt := product.UpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := product.ChangeDescription(tt.newDescription)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ChangeDescription() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("ChangeDescription() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ChangeDescription() unexpected error = %v", err)
				return
			}

			if product.Description() != tt.newDescription {
				t.Errorf("ChangeDescription() description = %v, want %v", product.Description(), tt.newDescription)
			}

			if !product.UpdatedAt().After(oldUpdatedAt) {
				t.Error("ChangeDescription() should update updatedAt timestamp")
			}
		})
	}
}

func TestProduct_ChangePrice(t *testing.T) {
	validProductType, _ := NewProductType(ProductTypeConsumable)

	tests := []struct {
		name     string
		newPrice int
		wantErr  error
	}{
		{
			name:     "valid price change",
			newPrice: 2000,
			wantErr:  nil,
		},
		{
			name:     "valid minimum price",
			newPrice: 1,
			wantErr:  nil,
		},
		{
			name:     "valid large price",
			newPrice: 999999999,
			wantErr:  nil,
		},
		{
			name:     "invalid price - zero",
			newPrice: 0,
			wantErr:  ErrInvalidPrice,
		},
		{
			name:     "invalid price - negative",
			newPrice: -100,
			wantErr:  ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("Product", "Description", 1000, validProductType)
			oldUpdatedAt := product.UpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := product.ChangePrice(tt.newPrice)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ChangePrice() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("ChangePrice() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ChangePrice() unexpected error = %v", err)
				return
			}

			if product.Price() != tt.newPrice {
				t.Errorf("ChangePrice() price = %v, want %v", product.Price(), tt.newPrice)
			}

			if !product.UpdatedAt().After(oldUpdatedAt) {
				t.Error("ChangePrice() should update updatedAt timestamp")
			}
		})
	}
}

func TestProduct_ChangeProductType(t *testing.T) {
	consumableType, _ := NewProductType(ProductTypeConsumable)
	simpleType, _ := NewProductType(ProductTypeSimple)
	partsType, _ := NewProductType(ProductTypeParts)

	tests := []struct {
		name         string
		initialType  ProductType
		newType      ProductType
		expectedType string
	}{
		{
			name:         "change from consumable to simple",
			initialType:  consumableType,
			newType:      simpleType,
			expectedType: ProductTypeSimple,
		},
		{
			name:         "change from simple to parts",
			initialType:  simpleType,
			newType:      partsType,
			expectedType: ProductTypeParts,
		},
		{
			name:         "change from parts to consumable",
			initialType:  partsType,
			newType:      consumableType,
			expectedType: ProductTypeConsumable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("Product", "Description", 1000, tt.initialType)
			oldUpdatedAt := product.UpdatedAt()
			time.Sleep(1 * time.Millisecond)

			product.ChangeProductType(tt.newType)

			if product.ProductType().Value() != tt.expectedType {
				t.Errorf("ChangeProductType() type = %v, want %v", product.ProductType().Value(), tt.expectedType)
			}

			if !product.UpdatedAt().After(oldUpdatedAt) {
				t.Error("ChangeProductType() should update updatedAt timestamp")
			}
		})
	}
}

func TestProduct_MarkAsDeleted(t *testing.T) {
	validProductType, _ := NewProductType(ProductTypeConsumable)
	product, _ := NewProduct("Product", "Description", 1000, validProductType)

	if product.IsDeleted() {
		t.Error("Product should not be deleted initially")
	}

	if product.DeletedAt() != nil {
		t.Error("DeletedAt should be nil initially")
	}

	oldUpdatedAt := product.UpdatedAt()
	time.Sleep(1 * time.Millisecond)

	product.MarkAsDeleted()

	if !product.IsDeleted() {
		t.Error("Product should be marked as deleted")
	}

	if product.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil after marking as deleted")
	}

	if !product.UpdatedAt().After(oldUpdatedAt) {
		t.Error("MarkAsDeleted() should update updatedAt timestamp")
	}

	if !product.DeletedAt().After(oldUpdatedAt) {
		t.Error("DeletedAt should be after the original updatedAt")
	}
}

func TestProduct_Getters(t *testing.T) {
	validUUID := utils.GenerateUUIDv7()
	validProductType, _ := NewProductType(ProductTypeParts)
	now := time.Now()
	deletedTime := time.Now().Add(-1 * time.Hour)

	product, _ := ReconstructProduct(
		validUUID,
		"Test Product",
		"Test Description",
		5000,
		validProductType,
		now,
		now,
		&deletedTime,
	)

	t.Run("ID getter", func(t *testing.T) {
		if product.ID() != validUUID {
			t.Errorf("ID() = %v, want %v", product.ID(), validUUID)
		}
	})

	t.Run("Name getter", func(t *testing.T) {
		if product.Name() != "Test Product" {
			t.Errorf("Name() = %v, want %v", product.Name(), "Test Product")
		}
	})

	t.Run("Description getter", func(t *testing.T) {
		if product.Description() != "Test Description" {
			t.Errorf("Description() = %v, want %v", product.Description(), "Test Description")
		}
	})

	t.Run("Price getter", func(t *testing.T) {
		if product.Price() != 5000 {
			t.Errorf("Price() = %v, want %v", product.Price(), 5000)
		}
	})

	t.Run("ProductType getter", func(t *testing.T) {
		if product.ProductType().Value() != ProductTypeParts {
			t.Errorf("ProductType() = %v, want %v", product.ProductType().Value(), ProductTypeParts)
		}
	})

	t.Run("CreatedAt getter", func(t *testing.T) {
		if !product.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", product.CreatedAt(), now)
		}
	})

	t.Run("UpdatedAt getter", func(t *testing.T) {
		if !product.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", product.UpdatedAt(), now)
		}
	})

	t.Run("DeletedAt getter", func(t *testing.T) {
		if product.DeletedAt() == nil {
			t.Error("DeletedAt() should not be nil")
		} else if !product.DeletedAt().Equal(deletedTime) {
			t.Errorf("DeletedAt() = %v, want %v", product.DeletedAt(), deletedTime)
		}
	})

	t.Run("IsDeleted getter", func(t *testing.T) {
		if !product.IsDeleted() {
			t.Error("IsDeleted() should return true")
		}
	})
}
