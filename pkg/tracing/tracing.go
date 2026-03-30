// Package tracing implements OpenTelemetry (OTel) setup and helper functions
// for distributed tracing across Gin HTTP servers, GORM, and outbound HTTP calls.
package tracing

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer initializes the OpenTelemetry tracer provider for a service
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Configure the OTLP HTTP exporter to send traces to localhost:4318 (OpenTelemetry Collector)
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(), // no TLS
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create a resource describing this application
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create the TracerProvider with the exporter and resource
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces
	)

	// Set the global TracerProvider
	otel.SetTracerProvider(tp)

	// Set the global TextMapPropagator for context propagation over HTTP
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("Tracer initialized", "service", serviceName)

	return tp, nil
}
