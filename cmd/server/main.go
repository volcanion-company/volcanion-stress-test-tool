package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/api/handler"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/api/router"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/engine"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.LogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Volcanion Stress Test Tool",
		zap.String("version", "1.0.0"),
		zap.String("port", cfg.ServerPort))

	// Initialize Prometheus metrics collector
	metricsCollector := metrics.NewCollector()
	logger.Log.Info("Metrics collector initialized")

	// Initialize repositories
	testPlanRepo := repository.NewMemoryTestPlanRepository()
	testRunRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	logger.Log.Info("Repositories initialized")

	// Initialize load generator
	loadGenerator := engine.NewLoadGenerator()
	logger.Log.Info("Load generator initialized")

	// Initialize service
	testService := service.NewTestService(testPlanRepo, testRunRepo, metricsRepo, loadGenerator)
	logger.Log.Info("Test service initialized")

	// Initialize handlers
	testPlanHandler := handler.NewTestPlanHandler(testService)
	testRunHandler := handler.NewTestRunHandler(testService)

	// Setup router
	r := router.SetupRouter(testPlanHandler, testRunHandler)
	logger.Log.Info("Router configured")

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Log.Info("Server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	logger.Log.Info("Server is ready to handle requests",
		zap.String("port", cfg.ServerPort))

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited")

	// Suppress unused variable warning
	_ = metricsCollector
}
