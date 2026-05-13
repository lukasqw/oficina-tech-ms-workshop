package handlers

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductHandler_UpdateProduct(t *testing.T) {
	tests := []struct {
		name                string
		productID           string
		requestBody         string
		mockProductRepo     *mockProductRepository
		mockInventoryRepo   *mockInventoryRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name:      "success - product updated",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{
				"name": "Updated Product",
				"description": "Updated Description",
				"price": 15000,
				"product_type": "SIMPLE"
			}`,
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
				saveFunc: func(_ context.Context, p *product.Product) error {
					return nil
				},
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"message":"Product updated successfully"`,
		},
		{
			name:                "error - invalid product ID format",
			productID:           "invalid-uuid",
			requestBody:         `{}`,
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_UUID"`,
		},
		{
			name:                "error - invalid request body",
			productID:           "550e8400-e29b-41d4-a716-446655440000",
			requestBody:         `{invalid json}`,
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_REQUEST"`,
		},
		{
			name:      "error - validation failed - empty name",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{
				"name": "",
				"description": "Updated Description",
				"price": 15000,
				"product_type": "SIMPLE"
			}`,
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
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"VALIDATION_FAILED"`,
		},
		{
			name:      "error - validation failed - invalid price",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{
				"name": "Updated Product",
				"description": "Updated Description",
				"price": 0,
				"product_type": "SIMPLE"
			}`,
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
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"VALIDATION_FAILED"`,
		},
		{
			name:      "error - product not found",
			productID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: `{
				"name": "Updated Product",
				"description": "Updated Description",
				"price": 15000,
				"product_type": "SIMPLE"
			}`,
			mockProductRepo: &mockProductRepository{
				findByIDFunc: func(_ context.Context, id string) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusNotFound,
			expectedBodyContain: `"code":"NOT_FOUND"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewProductHandler(tt.mockProductRepo, tt.mockInventoryRepo)

			req := httptest.NewRequest(http.MethodPut, "/products/"+tt.productID, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tt.productID)
			w := httptest.NewRecorder()

			handler.UpdateProduct(w, req)

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
