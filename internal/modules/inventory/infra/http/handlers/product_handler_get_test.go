package handlers

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductHandler_GetProductByID(t *testing.T) {
	tests := []struct {
		name                string
		productID           string
		mockProductRepo     *mockProductRepository
		mockInventoryRepo   *mockInventoryRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name:      "success - product found",
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
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"id":"550e8400-e29b-41d4-a716-446655440000"`,
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

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.productID, nil)
			req.SetPathValue("id", tt.productID)
			w := httptest.NewRecorder()

			handler.GetProductByID(w, req)

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

func TestProductHandler_GetAllProducts(t *testing.T) {
	tests := []struct {
		name                string
		mockProductRepo     *mockProductRepository
		mockInventoryRepo   *mockInventoryRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name: "success - products found",
			mockProductRepo: &mockProductRepository{
				findAllFunc: func(_ context.Context) ([]*product.Product, error) {
					pt, _ := product.NewProductType(product.ProductTypeSimple)
					return []*product.Product{
						createTestProduct(
							"550e8400-e29b-41d4-a716-446655440000",
							"Product 1",
							"Description 1",
							10000,
							pt,
						),
						createTestProduct(
							"550e8400-e29b-41d4-a716-446655440001",
							"Product 2",
							"Description 2",
							20000,
							pt,
						),
					}, nil
				},
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `"name":"Product 1"`,
		},
		{
			name: "success - empty list",
			mockProductRepo: &mockProductRepository{
				findAllFunc: func(_ context.Context) ([]*product.Product, error) {
					return []*product.Product{}, nil
				},
			},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusOK,
			expectedBodyContain: `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewProductHandler(tt.mockProductRepo, tt.mockInventoryRepo)

			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			w := httptest.NewRecorder()

			handler.GetAllProducts(w, req)

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
