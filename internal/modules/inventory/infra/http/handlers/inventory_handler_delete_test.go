package handlers

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInventoryHandler_DeleteInventory(t *testing.T) {
	tests := []struct {
		name                string
		productID           string
		mockInventoryRepo   *mockInventoryRepository
		mockProductRepo     *mockProductRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name:      "success - inventory deleted",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv := createTestInventory("660e8400-e29b-41d4-a716-446655440001", productID, 0, 0, 0)
					return inv, nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"message":"Inventory deleted successfully"`,
		},
		{
			name:                "error - invalid product ID format",
			productID:           "invalid-uuid",
			mockInventoryRepo:   &mockInventoryRepository{},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_UUID"`,
		},
		{
			name:      "error - inventory not found on get",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusNotFound,
			expectedBodyContain: `"code":"NOT_FOUND"`,
		},
		{
			name:      "error - inventory not found on delete",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv := createTestInventory("660e8400-e29b-41d4-a716-446655440001", productID, 0, 0, 0)
					return inv, nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return inventory.ErrInventoryNotFound
				},
			},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusNotFound,
			expectedBodyContain: `"code":"NOT_FOUND"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewInventoryHandler(tt.mockInventoryRepo, tt.mockProductRepo)

			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.productID+"/inventory", nil)
			req.SetPathValue("product_id", tt.productID)
			w := httptest.NewRecorder()

			handler.DeleteInventory(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if tt.expectedBodyContain != "" && !contains(body, tt.expectedBodyContain) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBodyContain, body)
			}
		})
	}
}
