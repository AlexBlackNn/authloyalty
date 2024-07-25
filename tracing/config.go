package tracing

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"sso/internal/config"
)

// // Init configures an OpenTelemetry exporter and trace provider.
//func Init() (*sdktrace.TracerProvider, error) {
//	exporter, err := stdout.New(stdout.WithPrettyPrint())
//	if err != nil {
//		return nil, err
//	}
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithSampler(sdktrace.AlwaysSample()),
//		sdktrace.WithBatcher(exporter),
//	)
//	otel.SetTracerProvider(tp)
//	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
//	return tp, nil
//}

func newResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("application", "otel-otlp-go-app"),
		),
	)
}

// Init configures an OpenTelemetry exporter and trace provider.
func Init(serviceName string, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	url := cfg.JaegerUrl
	jaegerExp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ResourceServiceName, err := newResource(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(ResourceServiceName),
		sdktrace.WithBatcher(jaegerExp),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}
