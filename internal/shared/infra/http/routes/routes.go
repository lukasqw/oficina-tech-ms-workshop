package routes

import (
	"net/http"
	"os"
	"time"

	inventoryHttp "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/infra/http"
	inventoryHandlers "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/infra/http/handlers"
	inventoryPersistence "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/infra/persistence"
	serviceCatalogHttp "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/infra/http"
	serviceCatalogHandlers "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/infra/http/handlers"
	serviceCatalogPersistence "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/infra/persistence"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/auth"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/database"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/http/middleware"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
)

func SetupRoutes() *http.ServeMux {
	jwtService := auth.NewJWTService()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	rbacMiddleware := middleware.NewRBACMiddleware()

	serviceRepository := serviceCatalogPersistence.NewServiceRepository(database.DB)
	serviceHandler := serviceCatalogHandlers.NewServiceHandler(serviceRepository)

	productRepository := inventoryPersistence.NewProductRepository(database.DB)
	inventoryRepository := inventoryPersistence.NewInventoryRepository(database.DB)
	productHandler := inventoryHandlers.NewProductHandler(productRepository, inventoryRepository)
	inventoryHandler := inventoryHandlers.NewInventoryHandler(inventoryRepository, productRepository)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	serviceCatalogHttp.RegisterServiceRoutes(mux, serviceHandler, authMiddleware, rbacMiddleware)
	inventoryHttp.RegisterInventoryRoutes(mux, productHandler, inventoryHandler, authMiddleware, rbacMiddleware)

	return mux
}

var startTime = time.Now()

func healthCheck(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	otelStatus := "inactive"
	if observability.OTelInitialized() {
		otelStatus = "active"
	}

	if database.DB == nil {
		utils.RespondSuccess(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "unhealthy",
			"database": "not initialized",
			"version":  version,
			"uptime":   time.Since(startTime).String(),
			"otel":     otelStatus,
		})
		return
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		utils.RespondSuccess(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "unhealthy",
			"database": "error",
			"version":  version,
			"uptime":   time.Since(startTime).String(),
			"otel":     otelStatus,
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		utils.RespondSuccess(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "unhealthy",
			"database": "disconnected",
			"version":  version,
			"uptime":   time.Since(startTime).String(),
			"otel":     otelStatus,
		})
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
		"status":   "healthy",
		"database": "connected",
		"version":  version,
		"uptime":   time.Since(startTime).String(),
		"otel":     otelStatus,
	})
}
