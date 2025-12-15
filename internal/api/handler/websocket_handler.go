package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"go.uber.org/zap"
)

// WebSocketHandler handles WebSocket connections for live metrics
type WebSocketHandler struct {
	testService *service.TestService
	logger      *zap.Logger
	upgrader    websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler with origin validation
func NewWebSocketHandler(testService *service.TestService, logger *zap.Logger, cfg *config.Config) *WebSocketHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Allow requests without Origin header (e.g., from non-browser clients)
				return true
			}
			// Validate against configured allowed origins
			return cfg.IsWebSocketOriginAllowed(origin)
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return &WebSocketHandler{
		testService: testService,
		logger:      logger,
		upgrader:    upgrader,
	}
}

// NewWebSocketHandlerPermissive creates a WebSocket handler that allows all origins (for development only)
// WARNING: Do not use in production!
func NewWebSocketHandlerPermissive(testService *service.TestService, logger *zap.Logger) *WebSocketHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true // Allow all origins - DEVELOPMENT ONLY
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return &WebSocketHandler{
		testService: testService,
		logger:      logger,
		upgrader:    upgrader,
	}
}

// LiveTestMetrics streams live metrics for a test run via WebSocket
func (h *WebSocketHandler) LiveTestMetrics(c *gin.Context) {
	runID := c.Param("id")

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established", zap.String("run_id", runID))

	// Send metrics every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Channel to handle client disconnection
	done := make(chan struct{})

	// Read from client (to detect disconnection)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(done)
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			h.logger.Info("WebSocket connection closed by client", zap.String("run_id", runID))
			return

		case <-ticker.C:
			// Get live metrics
			metrics, err := h.testService.GetLiveMetrics(runID)
			if err != nil {
				h.logger.Error("Failed to get live metrics", zap.Error(err))

				// Send error message to client
				errorMsg := map[string]string{
					"error": "Failed to get metrics",
				}
				if err := conn.WriteJSON(errorMsg); err != nil {
					h.logger.Error("Failed to send error message", zap.Error(err))
					return
				}
				continue
			}

			// Send metrics to client
			if err := conn.WriteJSON(metrics); err != nil {
				h.logger.Error("Failed to send metrics", zap.Error(err))
				return
			}

			// Check if test is complete
			testRun, err := h.testService.GetTestRun(runID)
			if err == nil && testRun.Status != "running" {
				h.logger.Info("Test completed, closing WebSocket", zap.String("run_id", runID))

				// Send final message
				finalMsg := map[string]interface{}{
					"status":  "completed",
					"metrics": metrics,
				}
				if err := conn.WriteJSON(finalMsg); err != nil {
					h.logger.Error("Failed to send final message", zap.Error(err))
				}
				return
			}
		}
	}
}

// LiveTestStatus streams test status updates via WebSocket
func (h *WebSocketHandler) LiveTestStatus(c *gin.Context) {
	runID := c.Param("id")

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(done)
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return

		case <-ticker.C:
			testRun, err := h.testService.GetTestRun(runID)
			if err != nil {
				continue
			}

			statusMsg := map[string]interface{}{
				"id":          testRun.ID,
				"status":      testRun.Status,
				"stop_reason": testRun.StopReason,
				"start_at":    testRun.StartAt,
				"end_at":      testRun.EndAt,
			}

			if err := conn.WriteJSON(statusMsg); err != nil {
				return
			}

			if testRun.Status != "running" {
				return
			}
		}
	}
}
