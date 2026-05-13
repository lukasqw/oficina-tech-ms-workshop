package http

import (
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/infra/http/handlers"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/http/middleware"
	"net/http"
)

// RegisterServiceRoutes registers all service catalog routes with authentication and authorization
func RegisterServiceRoutes(mux *http.ServeMux, handler *handlers.ServiceHandler, authMiddleware *middleware.AuthMiddleware, rbacMiddleware *middleware.RBACMiddleware) {
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

	// Register service routes with appropriate middleware chains
	mux.Handle("POST /services", writeMiddlewareChain(http.HandlerFunc(handler.CreateService)))
	mux.Handle("GET /services", readMiddlewareChain(http.HandlerFunc(handler.GetAllServices)))
	mux.Handle("GET /services/{id}", readMiddlewareChain(http.HandlerFunc(handler.GetServiceByID)))
	mux.Handle("PUT /services/{id}", writeMiddlewareChain(http.HandlerFunc(handler.UpdateService)))
	mux.Handle("DELETE /services/{id}", writeMiddlewareChain(http.HandlerFunc(handler.DeleteService)))
}
