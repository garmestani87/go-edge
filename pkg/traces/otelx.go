package traces

import (
	"context"
	"edge-app/configs"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(cfg *configs.Config) func(context.Context) error {
	var secureOption otlptracegrpc.Option
	var headerOption otlptracegrpc.Option

	if strings.ToLower(cfg.Otel.Insecure) == "false" {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		headerOption = otlptracegrpc.WithHeaders(map[string]string{
			"Authorization": cfg.Otel.BearerToken,
		})
	} else {
		headerOption = otlptracegrpc.WithHeaders(map[string]string{})
		secureOption = otlptracegrpc.WithInsecure()
	}
	// Sets up OTLP GRPC exporter with endpoint, headers, and TLS config.
	collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			headerOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	// Defines resource with service name, version, and environment.
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", cfg.Otel.ServiceName),
			attribute.String("service.version", cfg.Otel.ServiceVersion),
			attribute.String("language", cfg.Otel.Language),
			attribute.String("environment", cfg.Otel.DeploymentEnvironment),
		),
	)
	if err != nil {
		log.Fatalf("Could not set resources: %v", err)
	}

	// Configures the tracer provider with the exporter and resource.
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(traceProvider)

	// Sets global propagator to W3C Trace Context and Baggage.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return exporter.Shutdown
}
