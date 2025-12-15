package router

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/api/handler"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/middleware"
	"go.uber.org/zap"
)

//go:embed swagger-ui
var swaggerUI embed.FS

// RouterConfig holds configuration for router setup
//
//nolint:revive // exported name intentionally includes package name for clarity
type RouterConfig struct {
	TestPlanHandler     *handler.TestPlanHandler
	TestRunHandler      *handler.TestRunHandler
	ScenarioHandler     *handler.ScenarioHandler
	ReportHandler       *handler.ReportHandler
	WebSocketHandler    *handler.WebSocketHandler
	AuthHandler         *handler.AuthHandler
	AuditHandler        *handler.AuditHandler
	JWTService          *auth.JWTService
	APIKeyService       *auth.APIKeyService
	AuditMiddleware     gin.HandlerFunc
	RateLimitMiddleware gin.HandlerFunc
	AuthEnabled         bool
	Config              *config.Config
	Logger              *zap.Logger
	MetricsCollector    *metrics.Collector
	TracingEnabled      bool
}

// SetupRouter configures all API routes
func SetupRouter(routerConfig RouterConfig) *gin.Engine {
	// Use release mode for production (set GIN_MODE=release in production)
	// Can also be set via environment variable GIN_MODE
	if routerConfig.Config != nil && routerConfig.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New() // Use gin.New() instead of gin.Default() for custom middleware

	// Request ID middleware (always first)
	r.Use(middleware.RequestIDMiddleware())

	// Add custom recovery middleware with logging
	if routerConfig.Logger != nil {
		r.Use(middleware.RecoveryMiddleware(routerConfig.Logger))
	} else {
		r.Use(gin.Recovery()) // Fallback to default recovery
	}

	// Structured logging middleware (replaces gin.Logger())
	if routerConfig.Logger != nil {
		r.Use(middleware.LoggingMiddlewareWithConfig(middleware.LoggingConfig{
			Logger:    routerConfig.Logger,
			SkipPaths: []string{"/health", "/metrics"},
		}))
	} else {
		r.Use(gin.Logger()) // Fallback to default logger
	}

	// Metrics middleware
	if routerConfig.MetricsCollector != nil {
		r.Use(middleware.MetricsMiddlewareWithConfig(middleware.MetricsMiddlewareConfig{
			Collector: routerConfig.MetricsCollector,
			SkipPaths: []string{"/metrics"},
		}))
	}

	// Tracing middleware
	if routerConfig.TracingEnabled {
		r.Use(middleware.TracingMiddlewareWithConfig(middleware.TracingMiddlewareConfig{
			ServiceName: "volcanion-stress-test-tool",
			SkipPaths:   []string{"/health", "/metrics"},
		}))
	}

	// CORS middleware with config (apply globally)
	if routerConfig.Config != nil {
		r.Use(middleware.CORSMiddleware(routerConfig.Config))
	} else {
		// Fallback to permissive CORS for development
		r.Use(middleware.CORSMiddlewarePermissive())
	}

	// Apply audit middleware globally if provided
	if routerConfig.AuditMiddleware != nil {
		r.Use(routerConfig.AuditMiddleware)
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "volcanion-stress-test-tool",
		})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger UI and OpenAPI spec
	r.GET("/api/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/docs/")
	})
	// Serve Swagger UI static files
	swaggerFS, _ := fs.Sub(swaggerUI, "swagger-ui")
	r.StaticFS("/api/docs/", http.FS(swaggerFS))

	// Serve OpenAPI spec directly
	r.GET("/api/openapi.yaml", func(c *gin.Context) {
		c.File("./docs/openapi.yaml")
	})
	r.GET("/api/openapi.json", func(c *gin.Context) {
		c.File("./docs/openapi.json")
	})

	// API version header middleware
	apiVersionMiddleware := func(c *gin.Context) {
		c.Header("X-API-Version", "v1")
		c.Header("X-Deprecation-Notice", "") // Empty means not deprecated
		c.Next()
	}

	// API v1 routes
	api := r.Group("/api/v1")
	api.Use(apiVersionMiddleware)
	{
		// Public auth endpoints (no auth required)
		if routerConfig.AuthHandler != nil {
			authGroup := api.Group("/auth")
			{
				authGroup.POST("/login", routerConfig.AuthHandler.Login)
			}
		}

		// Protected API endpoints
		protected := api
		if routerConfig.AuthEnabled && routerConfig.JWTService != nil && routerConfig.APIKeyService != nil {
			// Apply authentication middleware
			protected.Use(middleware.AuthMiddleware(routerConfig.JWTService, routerConfig.APIKeyService))
		}

		// Apply rate limiting if provided
		if routerConfig.RateLimitMiddleware != nil {
			protected.Use(routerConfig.RateLimitMiddleware)
		}

		// Auth management endpoints (requires authentication)
		if routerConfig.AuthHandler != nil {
			authManagement := protected.Group("/auth")
			{
				authManagement.POST("/api-keys", routerConfig.AuthHandler.CreateAPIKey)
				authManagement.GET("/api-keys", routerConfig.AuthHandler.ListAPIKeys)
				authManagement.DELETE("/api-keys/:id", routerConfig.AuthHandler.RevokeAPIKey)
			}
		}

		// Audit endpoints (admin only)
		if routerConfig.AuditHandler != nil {
			audit := protected.Group("/audit")
			if routerConfig.AuthEnabled {
				audit.Use(middleware.RequireRole(auth.RoleAdmin))
			}
			{
				audit.GET("/logs", routerConfig.AuditHandler.GetAuditLogs)
				audit.GET("/export", routerConfig.AuditHandler.ExportAuditLogs)
			}
		}

		// Test Plan endpoints
		testPlans := protected.Group("/test-plans")
		{
			testPlans.POST("", routerConfig.TestPlanHandler.CreateTestPlan)
			testPlans.GET("", routerConfig.TestPlanHandler.GetTestPlans)
			testPlans.GET("/:id", routerConfig.TestPlanHandler.GetTestPlan)
		}

		// Test Run endpoints
		testRuns := protected.Group("/test-runs")
		{
			testRuns.POST("/start", routerConfig.TestRunHandler.StartTest)
			testRuns.POST("/:id/stop", routerConfig.TestRunHandler.StopTest)
			testRuns.GET("", routerConfig.TestRunHandler.GetTestRuns)
			testRuns.GET("/:id", routerConfig.TestRunHandler.GetTestRun)
			testRuns.GET("/:id/metrics", routerConfig.TestRunHandler.GetTestMetrics)
			testRuns.GET("/:id/live", routerConfig.TestRunHandler.GetLiveMetrics)

			// WebSocket endpoints for live updates
			if routerConfig.WebSocketHandler != nil {
				testRuns.GET("/:id/ws/metrics", routerConfig.WebSocketHandler.LiveTestMetrics)
				testRuns.GET("/:id/ws/status", routerConfig.WebSocketHandler.LiveTestStatus)
			}
		}

		// Scenario endpoints
		if routerConfig.ScenarioHandler != nil {
			scenarios := protected.Group("/scenarios")
			{
				scenarios.POST("", routerConfig.ScenarioHandler.CreateScenario)
				scenarios.GET("", routerConfig.ScenarioHandler.GetAllScenarios)
				scenarios.GET("/:id", routerConfig.ScenarioHandler.GetScenario)
				scenarios.DELETE("/:id", routerConfig.ScenarioHandler.DeleteScenario)
				scenarios.POST("/execute", routerConfig.ScenarioHandler.ExecuteScenario)
				scenarios.GET("/:id/executions", routerConfig.ScenarioHandler.GetScenarioExecutions)
				scenarios.GET("/executions/:id", routerConfig.ScenarioHandler.GetScenarioExecution)
			}
		}

		// Report endpoints
		if routerConfig.ReportHandler != nil {
			reports := protected.Group("/reports")
			{
				// Export test run in various formats
				reports.GET("/test-runs/:id/export", routerConfig.ReportHandler.ExportTestRun)

				// Compare test runs
				reports.POST("/compare", routerConfig.ReportHandler.CompareTestRuns)

				// SLA checking
				reports.POST("/test-runs/:id/sla", routerConfig.ReportHandler.CheckSLA)

				// Shareable reports
				reports.POST("/share", routerConfig.ReportHandler.CreateShareableReport)
				reports.GET("/shared", routerConfig.ReportHandler.ListSharedReports)
				reports.GET("/stats", routerConfig.ReportHandler.GetReportStats)
				reports.DELETE("/shared/:id", routerConfig.ReportHandler.DeleteSharedReport)
			}
		}
	}

	// Public shared report access (no auth required)
	if routerConfig.ReportHandler != nil {
		r.GET("/api/v1/reports/shared/:id", routerConfig.ReportHandler.GetSharedReport)
	}

	return r
}
