package handlers

import (
	"bytes"
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInventoryHandler_Reserve(t *testing.T) {
	tests := []struct {
		name                string
		productID           string
		requestBody         string
		mockInventoryRepo   *mockInventoryRepository
		mockProductRepo     *mockProductRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name:        "success - stock reserved",
			productID:   "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{"quantity": 10}`,
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv := createTestInventory("inv-123", productID, 100, 0, 0)
					return inv, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"reserved_quantity":10`,
		},
		{
			name:                "error - invalid product ID format",
			productID:           "invalid-uuid",
			requestBody:         `{"quantity": 10}`,
			mockInventoryRepo:   &mockInventoryRepository{},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_UUID"`,
		},
		{
			name:                "error - invalid request body",
			productID:           "550e8400-e29b-41d4-a716-446655440000",
			requestBody:         `invalid json`,
			mockInventoryRepo:   &mockInventoryRepository{},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_REQUEST"`,
		},
		{
			name:        "success - partial reserve with pending",
			productID:   "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{"quantity": 200}`,
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					inv := createTestInventory("inv-123", productID, 50, 0, 0)
					return inv, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			mockProductRepo:     &mockProductRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"pending_quantity":150`,
		},
		{
			name:        "error - inventory not found",
			productID:   "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{"quantity": 10}`,
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
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

			req := httptest.NewRequest(http.MethodPost, "/products/"+tt.productID+"/inventory/reserve", bytes.NewBufferString(tt.requestBody))
			req.SetPathValue("product_id", tt.productID)
			w := httptest.NewRecorder()

			handler.Reserve(w, req)

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
