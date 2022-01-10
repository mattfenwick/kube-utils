package simulator

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"math"
)

// NewJaegerTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func NewJaegerTracerProvider(service string, url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			//attribute.String("environment", environment),
			//attribute.Int64("ID", id),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func LittleJaegerExample(jaegerTracerProvider *tracesdk.TracerProvider, ctx context.Context) {
	mainTracer := jaegerTracerProvider.Tracer("main-tracer")

	ctx, span := mainTracer.Start(ctx, "main")
	defer span.End()

	// Use the global TracerProvider.
	subTracer := otel.Tracer("sub-tracer")
	_, coshSpan := subTracer.Start(ctx, "cosh")
	coshSpan.SetAttributes(attribute.Key("operation").String("my-cosh"))
	defer coshSpan.End()

	x := math.Cosh(67)
	logrus.Infof("cosh of 67: %+v", x)
}
