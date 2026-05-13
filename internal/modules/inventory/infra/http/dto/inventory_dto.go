package dto

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

// CreateInventoryRequest represents the HTTP request for creating an inventory
type CreateInventoryRequest struct {
	// Product ID to create inventory for (UUID format)
	ProductID string `json:"product_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// StockMovementRequest represents the HTTP request for stock movement operations
type StockMovementRequest struct {
	// Quantity to move (must be at least 1)
	Quantity int `json:"quantity" validate:"required,min=1" example:"10"`
}

// InventoryResponse represents the HTTP response for an inventory
type InventoryResponse struct {
	// Inventory unique identifier
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Product ID this inventory tracks
	ProductID string `json:"product_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Available stock quantity (ready to be reserved or used)
	AvailableQuantity int `json:"available_quantity" example:"50"`

	// Reserved stock quantity (allocated to service orders but not yet consumed)
	ReservedQuantity int `json:"reserved_quantity" example:"10"`

	// Pending stock quantity (confirmed usage, waiting to be removed from system)
	PendingQuantity int `json:"pending_quantity" example:"5"`

	// Creation timestamp in ISO 8601 format
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Last update timestamp in ISO 8601 format
	UpdatedAt string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ToInventoryResponse converts an inventory entity to an HTTP response
func ToInventoryResponse(inv *inventory.Inventory) *InventoryResponse {
	return &InventoryResponse{
		ID:                inv.ID(),
		ProductID:         inv.ProductID(),
		AvailableQuantity: inv.AvailableQuantity(),
		ReservedQuantity:  inv.ReservedQuantity(),
		PendingQuantity:   inv.PendingQuantity(),
		CreatedAt:         utils.FormatTimeRFC3339(inv.CreatedAt()),
		UpdatedAt:         utils.FormatTimeRFC3339(inv.UpdatedAt()),
	}
}
