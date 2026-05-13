package dto

// CreateServiceRequest represents the HTTP request for creating a service
type CreateServiceRequest struct {
	// Service name (3-100 characters)
	Name string `json:"name" validate:"required,min=3,max=100" example:"Troca de Óleo"`

	// Service description (10-500 characters)
	Description string `json:"description" validate:"required,min=10,max=500" example:"Troca de óleo do motor com filtro incluído"`

	// Service price in centavos (Brazilian Real). Example: 15000 = R$ 150,00
	Price int `json:"price" validate:"required,gt=0" example:"15000"`
}

// UpdateServiceRequest represents the HTTP request for updating a service
// All fields are optional to allow partial updates
type UpdateServiceRequest struct {
	// Service name (3-100 characters)
	Name *string `json:"name,omitempty" validate:"omitempty,min=3,max=100" example:"Troca de Óleo Completa"`

	// Service description (10-500 characters)
	Description *string `json:"description,omitempty" validate:"omitempty,min=10,max=500" example:"Troca de óleo do motor com filtro de óleo e ar incluídos"`

	// Service price in centavos (Brazilian Real). Example: 18000 = R$ 180,00
	Price *int `json:"price,omitempty" validate:"omitempty,gt=0" example:"18000"`
}

// ServiceResponse represents the HTTP response for a service
type ServiceResponse struct {
	// Service unique identifier
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Service name
	Name string `json:"name" example:"Troca de Óleo"`

	// Service description
	Description string `json:"description" example:"Troca de óleo do motor com filtro incluído"`

	// Service price in centavos (Brazilian Real). Example: 15000 = R$ 150,00
	Price int `json:"price" example:"15000"`

	// Creation timestamp in ISO 8601 format
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Last update timestamp in ISO 8601 format
	UpdatedAt string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}
