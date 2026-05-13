package http

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/infra/http/handlers"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/http/middleware"
	"net/http"
)

// RegisterInventoryRoutes registers all inventory-related routes with authentication and authorization
func RegisterInventoryRoutes(mux *http.ServeMux, productsHandler *handlers.ProductHandler, inventoryHandler *handlers.InventoryHandler, authMiddleware *middleware.AuthMiddleware, rbacMiddleware *middleware.RBACMiddleware) {
	// Middleware chain for read operations (USER, MANAGER, ADMIN)
	readMiddlewareChain := func(next http.Handler) http.Handler {
		return authMiddleware.Authenticate(
			rbacMiddleware.RequireRole("USER", "MANAGER", "ADMIN")(next))
	}

	// Middleware chain for write operations (MANAGER, ADMIN)
	writeMiddlewareChain := func(next http.Handler) http.Handler {
		return authMiddleware.Authenticate(
			rbacMiddleware.RequireRole("MANAGER", "ADMIN")(next))
	}

	// Product routes
	mux.Handle("POST /products", readMiddlewareChain(http.HandlerFunc(productsHandler.CreateProduct)))
	mux.Handle("GET /products", readMiddlewareChain(http.HandlerFunc(productsHandler.GetAllProducts)))
	mux.Handle("GET /products/{id}", readMiddlewareChain(http.HandlerFunc(productsHandler.GetProductByID)))
	mux.Handle("PUT /products/{id}", readMiddlewareChain(http.HandlerFunc(productsHandler.UpdateProduct)))
	mux.Handle("DELETE /products/{id}", readMiddlewareChain(http.HandlerFunc(productsHandler.DeleteProduct)))

	// Inventory routes - RESTful pattern using product ID
	mux.Handle("POST /products/{product_id}/inventory", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.CreateInventory)))
	mux.Handle("GET /products/{product_id}/inventory", readMiddlewareChain(http.HandlerFunc(inventoryHandler.GetInventoryByProduct)))
	mux.Handle("DELETE /products/{product_id}/inventory", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.DeleteInventory)))

	// Stock movement operations (all require MANAGER or ADMIN)
	mux.Handle("POST /products/{product_id}/inventory/manual-decrease", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.ManualDecrease)))
	mux.Handle("POST /products/{product_id}/inventory/reserved-decrease", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.ReservedDecrease)))
	mux.Handle("POST /products/{product_id}/inventory/available-decrease", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.AvailableDecrease)))
	mux.Handle("POST /products/{product_id}/inventory/reserve", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.Reserve)))
	mux.Handle("POST /products/{product_id}/inventory/increase", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.Increase)))
	mux.Handle("POST /products/{product_id}/inventory/cancel-reserved", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.CancelReserved)))
	mux.Handle("POST /products/{product_id}/inventory/cancel-confirmed", writeMiddlewareChain(http.HandlerFunc(inventoryHandler.CancelConfirmed)))
}
