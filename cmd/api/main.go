package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/messaging/consumers"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/messaging/publishers"
	inventoryUsecases "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
	inventoryPersistence "github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/infra/persistence"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/awsconfig"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/database"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/http/middleware"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/http/routes"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	sqsinfra "github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/sqs"
)

const (
	queueOrderInventoryRequested = "order-inventory-op-requested"
	queueOrderInventorySucceeded = "order-inventory-op-succeeded"
	queueOrderInventoryFailed    = "order-inventory-op-failed"
)

// @title           Oficina Tech Workshop API
// @version         1.0
// @description     REST API para gestão de oficina automotiva no Brasil

func main() {
	observability.NewLogger()

	if err := godotenv.Load(); err != nil {
		slog.Warn("Warning: .env file not found, using system environment variables")
	}

	startupCtx := context.Background()
	otelShutdown, err := observability.InitOTel(startupCtx)
	if err != nil {
		slog.Error("failed to initialize OpenTelemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			slog.Error("OTel shutdown error", "error", err)
		}
	}()

	database.Connect()

	mux := routes.SetupRoutes()
	handler := middleware.NewObservabilityMiddleware(middleware.WrapMux(mux))

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	if err := startSagaConsumer(workerCtx); err != nil {
		slog.Error("saga consumer initialization failed", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8083"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: handler,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		slog.Info("Received shutdown signal", "signal", sig.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		}
		cancelWorkers()
	}()

	slog.Info("Server starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}

func startSagaConsumer(ctx context.Context) error {
	if os.Getenv("AWS_ENDPOINT_URL") == "" && os.Getenv("AWS_REGION") == "" && os.Getenv("AWS_DEFAULT_REGION") == "" {
		slog.Warn("AWS not configured; skipping saga consumer")
		return nil
	}

	awsCfg, err := awsconfig.Load(ctx)
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}
	sqsClient := sqsinfra.NewClient(awsCfg)

	queueURLs, err := sqsinfra.ResolveQueueURLs(ctx, sqsClient,
		queueOrderInventoryRequested,
		queueOrderInventorySucceeded,
		queueOrderInventoryFailed,
	)
	if err != nil {
		return fmt.Errorf("resolve queue URLs: %w", err)
	}

	inventoryRepo := inventoryPersistence.NewInventoryRepository(database.DB)
	sagaRepo := inventoryPersistence.NewSagaOperationRepository(database.DB)
	useCase := inventoryUsecases.NewProcessSagaOperationUseCase(sagaRepo, inventoryRepo)

	publisher := publishers.NewOrderInventoryOperationPublisher(
		sqsClient,
		queueURLs[queueOrderInventorySucceeded],
		queueURLs[queueOrderInventoryFailed],
	)

	consumer := consumers.NewOrderInventoryOperationRequestedConsumer(
		sqsClient,
		queueURLs[queueOrderInventoryRequested],
		useCase,
		publisher,
	)

	go func() {
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			slog.Error("saga consumer stopped", "error", err)
		}
	}()
	slog.Info("saga consumer started", "queue", queueURLs[queueOrderInventoryRequested])
	return nil
}
