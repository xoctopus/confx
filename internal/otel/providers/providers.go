package providers

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xoctopus/x/contextx"
	otelapilogger "go.opentelemetry.io/otel/log"
	otelapimetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	otelsdklogger "go.opentelemetry.io/otel/sdk/log"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"
	otelsdktracer "go.opentelemetry.io/otel/sdk/trace"
	otelapitracer "go.opentelemetry.io/otel/trace"
)

type (
	_ = otelapilogger.LoggerProvider
	_ = otelapitracer.TracerProvider
	_ = otelapimetric.MeterProvider

	_ = otelsdklogger.LoggerProvider
	_ = otelsdktracer.TracerProvider
	_ = otelsdkmetric.MeterProvider

	tCtxLoggerProvider struct{}
	tCtxTracerProvider struct{}
	tCtxMetricProvider struct{}
	tCtxMetricGatherer struct{}
)

var (
	CtxGatherer = contextx.NewT[prometheus.Gatherer]

	LoggerProviderFrom  = contextx.From[tCtxLoggerProvider, otelapilogger.LoggerProvider]
	MustLoggerProvider  = contextx.Must[tCtxLoggerProvider, otelapilogger.LoggerProvider]
	WithLoggerProvider  = contextx.With[tCtxLoggerProvider, otelapilogger.LoggerProvider]
	CarryLoggerProvider = contextx.Carry[tCtxLoggerProvider, otelapilogger.LoggerProvider]

	TracerProviderFrom  = contextx.From[tCtxTracerProvider, otelapitracer.TracerProvider]
	MustTracerProvider  = contextx.Must[tCtxTracerProvider, otelapitracer.TracerProvider]
	WithTracerProvider  = contextx.With[tCtxTracerProvider, otelapitracer.TracerProvider]
	CarryTracerProvider = contextx.Carry[tCtxTracerProvider, otelapitracer.TracerProvider]

	MustMetricProvider  = contextx.Must[tCtxMetricProvider, otelapimetric.MeterProvider]
	WithMetricProvider  = contextx.With[tCtxMetricProvider, otelapimetric.MeterProvider]
	CarryMetricProvider = contextx.Carry[tCtxMetricProvider, otelapimetric.MeterProvider]

	MeterGathererFrom   = contextx.From[tCtxMetricGatherer, prometheus.Gatherer]
	MustMetricGatherer  = contextx.Must[tCtxMetricProvider, prometheus.Gatherer]
	WithMetricGatherer  = contextx.With[tCtxMetricProvider, prometheus.Gatherer]
	CarryMetricGatherer = contextx.Carry[tCtxMetricProvider, prometheus.Gatherer]
)

func MetricProviderFrom(ctx context.Context) otelapimetric.MeterProvider {
	return contextx.FromOr[tCtxMetricProvider](ctx, noop.NewMeterProvider())
}

func MeterFrom(ctx context.Context) otelapimetric.Meter {
	return MetricProviderFrom(ctx).Meter("")
}
