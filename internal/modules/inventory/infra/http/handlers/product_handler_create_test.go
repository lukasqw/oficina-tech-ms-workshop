package handlers

import (
	"context"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name                string
		requestBody         string
		mockProductRepo     *mockProductRepository
		mockInventoryRepo   *mockInventoryRepository
		expectedStatus      int
		expectedBodyContain string
	}{
		{
			name: "success - product created",
			requestBody: `{
				"name": "Test Product",
				"description": "Test Description",
				"price": 10000,
				"product_type": "SIMPLE"
			}`,
			mockProductRepo: &mockProductRepository{
				existsByNameFunc: func(_ context.Context, name string) (bool, error) {
					return false, nil
				},
				saveFunc: func(_ context.Context, p *product.Product) error {
					// Simulate setting ID after save
					return p.SetID("550e8400-e29b-41d4-a716-446655440000")
				},
			},
			mockInventoryRepo: &mockInventoryRepository{
				saveFunc: func(_ context.Context, inv *inventory.Inventory) error {
					return nil
				},
			},
			expectedStatus:      http.StatusCreated,
			expectedBodyContain: `"name":"Test Product"`,
		},
		{
			name:                "error - invalid request body",
			requestBody:         `{invalid json}`,
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"INVALID_REQUEST"`,
		},
		{
			name: "error - validation failed - empty name",
			requestBody: `{
				"name": "",
				"description": "Test Description",
				"price": 10000,
				"product_type": "SIMPLE"
			}`,
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"VALIDATION_FAILED"`,
		},
		{
			name: "error - validation failed - invalid price",
			requestBody: `{
				"name": "Test Product",
				"description": "Test Description",
				"price": 0,
				"product_type": "SIMPLE"
			}`,
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"VALIDATION_FAILED"`,
		},
		{
			name: "error - validation failed - invalid product type",
			requestBody: `{
				"name": "Test Product",
				"description": "Test Description",
				"price": 10000,
				"product_type": "INVALID"
			}`,
			mockProductRepo:     &mockProductRepository{},
			mockInventoryRepo:   &mockInventoryRepository{},
			expectedStatus:      http.StatusBadRequest,
			expectedBodyContain: `"code":"VALIDATION_FAILED"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewProductHandler(tt.mockProductRepo, tt.mockInventoryRepo)

			req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateProduct(w, req)

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
