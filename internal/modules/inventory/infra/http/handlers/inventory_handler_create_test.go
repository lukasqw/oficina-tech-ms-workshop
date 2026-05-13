package handlers

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInventoryHandler_CreateInventory(t *testing.T) {
	tests := []struct {
		name                string
		productID           string
		mockInventoryRepo   *mockInventoryRepository
		mockProductRepo     *mockProductRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name:      "success - inventory created",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockInventoryRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return false, nil
				},
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			mockProductRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					pt, _ := product.NewProductType(product.ProductTypeSimple)
					p, _ := product.NewProduct("Test Product", "Test Description", 100, pt)
					return p, nil
				},
			},
			expectedStatus:      http.StatusCreated,
			expectedBodyContain: `"product_id":"550e8400-e29b-41d4-a716-446655440000"`,
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
			name:      "error - product not found",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockInventoryRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return false, nil
				},
			},
			mockProductRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},
			expectedStatus:      http.StatusInternalServerError,
			expectedBodyContain: `"code":"INTERNAL_ERROR"`,
		},
		{
			name:      "error - product already has inventory",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockInventoryRepo: &mockInventoryRepository{
				existsByProductIDFunc: func(_ context.Context, productID string) (bool, error) {
					return true, nil
				},
			},
			mockProductRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					pt, _ := product.NewProductType(product.ProductTypeSimple)
					p, _ := product.NewProduct("Test Product", "Test Description", 100, pt)
					return p, nil
				},
			},
			expectedStatus:      http.StatusConflict,
			expectedBodyContain: `"code":"DUPLICATE_RESOURCE"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewInventoryHandler(tt.mockInventoryRepo, tt.mockProductRepo)

			req := httptest.NewRequest(http.MethodPost, "/products/"+tt.productID+"/inventory", nil)
			req.SetPathValue("product_id", tt.productID)
			w := httptest.NewRecorder()

			handler.CreateInventory(w, req)

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
