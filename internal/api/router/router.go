package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/api/handler"
)

// SetupRouter configures all API routes
func SetupRouter(testPlanHandler *handler.TestPlanHandler, testRunHandler *handler.TestRunHandler) *gin.Engine {
	// Use release mode for production
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "volcanion-stress-test-tool",
		})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	api := r.Group("/api")
	{
		// Test Plan endpoints
		testPlans := api.Group("/test-plans")
		{
			testPlans.POST("", testPlanHandler.CreateTestPlan)
			testPlans.GET("", testPlanHandler.GetTestPlans)
			testPlans.GET("/:id", testPlanHandler.GetTestPlan)
		}

		// Test Run endpoints
		testRuns := api.Group("/test-runs")
		{
			testRuns.POST("/start", testRunHandler.StartTest)
			testRuns.POST("/:id/stop", testRunHandler.StopTest)
			testRuns.GET("", testRunHandler.GetTestRuns)
			testRuns.GET("/:id", testRunHandler.GetTestRun)
			testRuns.GET("/:id/metrics", testRunHandler.GetTestMetrics)
			testRuns.GET("/:id/live", testRunHandler.GetLiveMetrics)
		}
	}

	return r
}
