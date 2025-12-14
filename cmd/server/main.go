package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/api/handler"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/api/router"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/audit"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/engine"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/middleware"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/reporting"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/postgres"
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

	// Initialize database connection (if configured)
	var db *sql.DB
	var metricsSnapshotRepo *postgres.MetricsSnapshotRepository

	if cfg.DatabaseDSN != "" {
		dbConn, err := postgres.NewDB(postgres.DBConfig{
			DSN:          cfg.DatabaseDSN,
			MaxConns:     cfg.DatabaseMaxConns,
			MaxIdleConns: cfg.DatabaseMaxIdleConns,
		}, logger.Log)

		if err != nil {
			logger.Log.Fatal("Failed to connect to database", zap.Error(err))
		}
		db = dbConn
		metricsSnapshotRepo = postgres.NewMetricsSnapshotRepository(db)
		logger.Log.Info("PostgreSQL database connected")

		defer func() {
			if err := db.Close(); err != nil {
				logger.Log.Error("Failed to close database", zap.Error(err))
			}
		}()
	} else {
		logger.Log.Warn("DATABASE_DSN not configured, using in-memory storage")
	}

	// Initialize repositories (PostgreSQL or in-memory)
	var testPlanRepo repository.TestPlanRepository
	var testRunRepo repository.TestRunRepository
	var metricsRepo repository.MetricsRepository

	scenarioRepo := repository.NewMemoryScenarioRepository()
	scenarioExecutionRepo := repository.NewMemoryScenarioExecutionRepository()

	if db != nil {
		testPlanRepo = postgres.NewPostgresTestPlanRepository(db)
		testRunRepo = postgres.NewPostgresTestRunRepository(db)
		metricsRepo = postgres.NewPostgresMetricsRepository(db)
		logger.Log.Info("Using PostgreSQL repositories")
	} else {
		testPlanRepo = repository.NewMemoryTestPlanRepository()
		testRunRepo = repository.NewMemoryTestRunRepository()
		metricsRepo = repository.NewMemoryMetricsRepository()
		logger.Log.Info("Using in-memory repositories")
	}

	// Initialize scenario executor
	scenarioExecutor := engine.NewScenarioExecutor()
	logger.Log.Info("Scenario executor initialized")

	// Initialize load generator with collector
	loadGenerator := engine.NewLoadGenerator(metricsCollector)
	logger.Log.Info("Load generator initialized")

	// Initialize service
	testService := service.NewTestService(testPlanRepo, testRunRepo, metricsRepo, loadGenerator, cfg)
	logger.Log.Info("Test service initialized")

	// Create individual services for report handler
	testPlanService := service.NewTestPlanService(testPlanRepo)
	testRunService := service.NewTestRunService(testRunRepo)
	metricsService := service.NewMetricsService(metricsRepo)

	scenarioService := service.NewScenarioService(scenarioRepo, scenarioExecutionRepo, scenarioExecutor)
	logger.Log.Info("Scenario service initialized")

	// Initialize auth services
	jwtService := auth.NewJWTService(cfg.JWTSecret, time.Duration(cfg.JWTDuration)*time.Hour)
	apiKeyService := auth.NewAPIKeyService()
	logger.Log.Info("Auth services initialized")

	// Initialize audit logger
	auditLogger := audit.NewLogger(logger.Log, 10000) // Keep last 10k events in memory
	logger.Log.Info("Audit logger initialized")

	// Initialize report store
	reportStore := reporting.NewReportStore()
	logger.Log.Info("Report store initialized")

	// Initialize handlers
	testPlanHandler := handler.NewTestPlanHandler(testService)
	testRunHandler := handler.NewTestRunHandler(testService)
	scenarioHandler := handler.NewScenarioHandler(scenarioService)
	authHandler := handler.NewAuthHandler(jwtService, apiKeyService)
	auditHandler := handler.NewAuditHandler(auditLogger)
	reportHandler := handler.NewReportHandler(
		testRunService,
		testPlanService,
		metricsService,
		reportStore,
	)
	websocketHandler := handler.NewWebSocketHandler(testService, logger.Log, cfg)

	var metricsHandler *handler.MetricsHandler
	if metricsSnapshotRepo != nil {
		metricsHandler = handler.NewMetricsHandler(metricsSnapshotRepo, logger.Log)
	}

	// Initialize middlewares
	auditMiddleware := middleware.AuditMiddleware(auditLogger)
	var rateLimitMiddleware gin.HandlerFunc
	if cfg.RateLimitEnabled {
		rateLimiter := middleware.NewPerUserRateLimiter(
			100.0,                  // admin: 100 req/s
			cfg.RateLimitPerSecond, // user: from config
			5.0,                    // readonly: 5 req/s
			2.0,                    // default/unauthenticated: 2 req/s
			20,                     // burst
		)
		rateLimitMiddleware = middleware.PerUserRateLimitMiddleware(rateLimiter)
		logger.Log.Info("Rate limiting enabled",
			zap.Float64("rate_per_second", cfg.RateLimitPerSecond))
	}

	// Setup router
	r := router.SetupRouter(router.RouterConfig{
		TestPlanHandler:     testPlanHandler,
		TestRunHandler:      testRunHandler,
		ScenarioHandler:     scenarioHandler,
		ReportHandler:       reportHandler,
		WebSocketHandler:    websocketHandler,
		AuthHandler:         authHandler,
		AuditHandler:        auditHandler,
		JWTService:          jwtService,
		APIKeyService:       apiKeyService,
		AuditMiddleware:     auditMiddleware,
		RateLimitMiddleware: rateLimitMiddleware,
		AuthEnabled:         cfg.AuthEnabled,
		Config:              cfg,
		Logger:              logger.Log,
	})

	// Add metrics endpoint if database is configured
	if metricsHandler != nil {
		r.GET("/api/metrics/history/:runId", metricsHandler.GetHistoricalMetrics)
		logger.Log.Info("Historical metrics endpoint enabled")
	}

	logger.Log.Info("Router configured",
		zap.Bool("auth_enabled", cfg.AuthEnabled),
		zap.Bool("rate_limit_enabled", cfg.RateLimitEnabled))

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

	// Stop all active tests before shutting down server
	if err := loadGenerator.Shutdown(20 * time.Second); err != nil {
		logger.Log.Error("Failed to gracefully shutdown tests", zap.Error(err))
	}

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
