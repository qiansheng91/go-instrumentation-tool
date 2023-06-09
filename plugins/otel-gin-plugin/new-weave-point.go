package otel_gin_plugin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func beforeNewMethod(parameters []interface{}) {
	exporter, _ := stdout.New(stdout.WithPrettyPrint())
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func afterNewMethod(ret []interface{}) {
	ginEngine := ret[0].(**gin.Engine)
	(*ginEngine).Use(middleWare)
}

func middleWare(c *gin.Context) {
	savedCtx := c.Request.Context()
	defer func() {
		c.Request = c.Request.WithContext(savedCtx)
	}()
	ctx := otel.GetTextMapPropagator().Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}
	tracer := otel.GetTracerProvider().Tracer(
		"test",
		oteltrace.WithInstrumentationVersion("1.0.0"),
	)

	var spanName = c.FullPath()
	ctx, span := tracer.Start(ctx, spanName, opts...)
	defer span.End()

	// pass the span through the request context
	c.Request = c.Request.WithContext(ctx)

	// serve the request to the next middleware
	c.Next()

	status := c.Writer.Status()
	if status > 0 {
		span.SetAttributes(semconv.HTTPStatusCode(status))
	}
	if len(c.Errors) > 0 {
		span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
	}
}
