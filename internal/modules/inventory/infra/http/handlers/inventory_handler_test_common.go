package handlers

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"time"
)

// Helper to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper to create test inventory
func createTestInventory(id, productID string, available, reserved, pending int) *inventory.Inventory {
	now := time.Now()
	inv, err := inventory.ReconstructInventory(id, productID, available, reserved, pending, now, now, nil)
	if err != nil {
		panic("failed to create test inventory: " + err.Error())
	}
	return inv
}

// Helper to create test product
func createTestProduct(id, name, description string, price int, productType product.ProductType) *product.Product {
	now := time.Now()
	p, err := product.ReconstructProduct(id, name, description, price, productType, now, now, nil)
	if err != nil {
		panic("failed to create test product: " + err.Error())
	}
	return p
}
