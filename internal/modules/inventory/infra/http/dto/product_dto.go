package dto

// CreateProductRequest represents the HTTP request for creating a product
type CreateProductRequest struct {
	// Product name (2-200 characters)
	Name string `json:"name" validate:"required,min=2,max=200" example:"Óleo de Motor 5W30"`

	// Product description (optional, max 1000 characters)
	Description string `json:"description" validate:"max=1000" example:"Óleo sintético para motores a gasolina"`

	// Product price in cents (must be greater than 0)
	Price int `json:"price" validate:"required,gt=0" example:"8500"`

	// Product type: CONSUMABLE (consumable items), SIMPLE (simple products), or PARTS (vehicle parts)
	ProductType string `json:"product_type" validate:"required,oneof=CONSUMABLE SIMPLE PARTS" example:"CONSUMABLE" enums:"CONSUMABLE,SIMPLE,PARTS"`
}

// UpdateProductRequest represents the HTTP request for updating a product
// All fields are optional to allow partial updates
type UpdateProductRequest struct {
	// Product name (2-200 characters)
	Name *string `json:"name,omitempty" validate:"omitempty,min=2,max=200" example:"Óleo de Motor 5W30"`

	// Product description (max 1000 characters)
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Óleo sintético para motores a gasolina"`

	// Product price in cents (must be greater than 0)
	Price *int `json:"price,omitempty" validate:"omitempty,gt=0" example:"8500"`

	// Product type: CONSUMABLE, SIMPLE, or PARTS
	ProductType *string `json:"product_type,omitempty" validate:"omitempty,oneof=CONSUMABLE SIMPLE PARTS" example:"CONSUMABLE" enums:"CONSUMABLE,SIMPLE,PARTS"`
}

// ProductResponse represents the HTTP response for a product
type ProductResponse struct {
	// Product unique identifier
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Product name
	Name string `json:"name" example:"Óleo de Motor 5W30"`

	// Product description
	Description string `json:"description" example:"Óleo sintético para motores a gasolina"`

	// Product price in cents
	Price int `json:"price" example:"8500"`

	// Product type
	ProductType string `json:"product_type" example:"CONSUMABLE" enums:"CONSUMABLE,SIMPLE,PARTS"`

	// Creation timestamp in ISO 8601 format
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Last update timestamp in ISO 8601 format
	UpdatedAt string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}
