package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig holds configuration for tracing
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string // OTLP endpoint (e.g., "localhost:4317")
	Insecure       bool   // Use insecure connection for OTLP
	SampleRate     float64
	Enabled        bool
	UseStdout      bool // For debugging - output traces to stdout
}

// DefaultTracerConfig returns default tracing configuration
func DefaultTracerConfig() TracerConfig {
	return TracerConfig{
		ServiceName:    "volcanion-stress-test-tool",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Endpoint:       "localhost:4317",
		Insecure:       true,
		SampleRate:     1.0, // Sample all traces in development
		Enabled:        false,
		UseStdout:      false,
	}
}

// TracerProvider wraps the OpenTelemetry TracerProvider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(config TracerConfig) (*TracerProvider, error) {
	if !config.Enabled {
		// Return a no-op tracer
		return &TracerProvider{
			tracer: otel.Tracer(config.ServiceName),
		}, nil
	}

	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter
	var exporter sdktrace.SpanExporter

	if config.UseStdout {
		// Stdout exporter for debugging
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
	} else {
		// OTLP exporter for production
		var opts []otlptracegrpc.Option
		opts = append(opts, otlptracegrpc.WithEndpoint(config.Endpoint))

		if config.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		client := otlptracegrpc.NewClient(opts...)
		exporter, err = otlptrace.New(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
	}

	// Create sampler
	var sampler sdktrace.Sampler
	switch {
	case config.SampleRate >= 1.0:
		sampler = sdktrace.AlwaysSample()
	case config.SampleRate <= 0:
		sampler = sdktrace.NeverSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(config.SampleRate)
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{
		provider: tp,
		tracer:   tp.Tracer(config.ServiceName),
	}, nil
}

// Tracer returns the configured tracer
func (tp *TracerProvider) Tracer() trace.Tracer {
	return tp.tracer
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		return tp.provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
