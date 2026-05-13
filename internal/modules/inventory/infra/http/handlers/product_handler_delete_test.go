package handlers

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductHandler_DeleteProduct(t *testing.T) {
	tests := []struct {
		name                string
		productID           string
		mockProductRepo     *mockProductRepository
		mockInventoryRepo   *mockInventoryRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name:      "success - product deleted",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockProductRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					pt, _ := product.NewProductType(product.ProductTypeSimple)
					return createTestProduct(
						"550e8400-e29b-41d4-a716-446655440000",
						"Test Product",
						"Test Description",
						10000,
						pt,
					), nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return createTestInventory(
						"550e8400-e29b-41d4-a716-446655440001",
						productID,
						0,
						0,
						0,
					), nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"message":"Product deleted successfully"`,
		},
		{
			name:                "error - invalid product ID format",
			productID:           "invalid-uuid",
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_UUID"`,
		},
		{
			name:      "error - product not found",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockProductRepo: &mockProductRepository{
				deleteFunc: func(_ context.Context, id string) error {
					return product.ErrProductNotFound
				},
			},
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return nil, inventory.ErrInventoryNotFound
				},
			},
			expectedStatus:      http.StatusNotFound,
			expectedBodyContain: `"code":"NOT_FOUND"`,
		},
		{
			name:      "success - product deleted even with available stock",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			mockProductRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					pt, _ := product.NewProductType(product.ProductTypeSimple)
					return createTestProduct(
						"550e8400-e29b-41d4-a716-446655440000",
						"Test Product",
						"Test Description",
						10000,
						pt,
					), nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			mockInventoryRepo: &mockInventoryRepository{
				findByProductIDFunc: func(_ context.Context, productID string) (*inventory.Inventory, error) {
					return createTestInventory(
						"550e8400-e29b-41d4-a716-446655440001",
						productID,
						10,
						0,
						0,
					), nil
				},
				deleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"message":"Product deleted successfully"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewProductHandler(tt.mockProductRepo, tt.mockInventoryRepo)

			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.productID, nil)
			req.SetPathValue("id", tt.productID)
			w := httptest.NewRecorder()

			handler.DeleteProduct(w, req)

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
