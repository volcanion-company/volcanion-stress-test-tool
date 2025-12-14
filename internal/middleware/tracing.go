package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/volcanion-company/volcanion-stress-test-tool/internal/middleware"

// TracingMiddleware adds OpenTelemetry tracing to HTTP requests
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract context from incoming request headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Get the route path for span name
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Start span
		spanName := c.Request.Method + " " + path
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Request.Method),
				semconv.HTTPRouteKey.String(path),
				semconv.URLPath(c.Request.URL.Path),
				semconv.URLScheme(c.Request.URL.Scheme),
				semconv.ServerAddress(c.Request.Host),
				semconv.UserAgentOriginal(c.Request.UserAgent()),
				semconv.ClientAddress(c.ClientIP()),
				attribute.String("http.request_id", GetRequestID(c)),
			),
		)
		defer span.End()

		// Set the context with span
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add response attributes
		statusCode := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPResponseStatusCode(statusCode),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Record errors
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("http.error", c.Errors.String()))
			for _, err := range c.Errors {
				span.RecordError(err.Err)
			}
		}

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// TracingMiddlewareConfig allows custom configuration
type TracingMiddlewareConfig struct {
	ServiceName string
	SkipPaths   []string
	Tracer      trace.Tracer
}

// TracingMiddlewareWithConfig creates a tracing middleware with custom config
func TracingMiddlewareWithConfig(config TracingMiddlewareConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	tracer := config.Tracer
	if tracer == nil {
		tracer = otel.Tracer(tracerName)
	}
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip tracing for specified paths
		if skipPaths[path] {
			c.Next()
			return
		}

		// Extract context from incoming request headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Get the route path for span name
		routePath := c.FullPath()
		if routePath == "" {
			routePath = path
		}

		// Start span
		spanName := c.Request.Method + " " + routePath
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Request.Method),
				semconv.HTTPRouteKey.String(routePath),
				semconv.URLPath(path),
				attribute.String("http.request_id", GetRequestID(c)),
			),
		)
		defer span.End()

		// Set the context with span
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add response attributes
		statusCode := c.Writer.Status()
		span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))

		// Record errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				span.RecordError(err.Err)
			}
		}
	}
}

// InjectTraceContext injects trace context into outgoing HTTP request headers
func InjectTraceContext(c *gin.Context, headers map[string]string) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}

	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier(headers)
	propagator.Inject(c.Request.Context(), carrier)

	return headers
}
