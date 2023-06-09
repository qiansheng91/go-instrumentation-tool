package otel_gin_plugin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func beforeNewMethod(obj interface{}, variables map[string]string, parameters []interface{}) {
	exporter, _ := stdout.New(stdout.WithPrettyPrint())
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func afterNewMethod(obj interface{}, variables map[string]string, parameters []interface{}, ret []interface{}) {
	ginEngine := ret[0].(**gin.Engine)
	(*ginEngine).Use(otelgin.Middleware("test-server"))
}
