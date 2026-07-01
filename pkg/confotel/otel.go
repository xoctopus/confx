package confotel

import (
	"context"
	"errors"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/contextx"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelpromexporter "go.opentelemetry.io/otel/exporters/prometheus"
	otelsdklogger "go.opentelemetry.io/otel/sdk/log"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"
	otelsdkresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktracer "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.org/x/sync/errgroup"

	"github.com/xoctopus/confx/internal/otel/exporter"
	"github.com/xoctopus/confx/internal/otel/logger"
	"github.com/xoctopus/confx/internal/otel/providers"
	"github.com/xoctopus/confx/pkg/confotel/metric"
	"github.com/xoctopus/confx/pkg/types"
)

type Otel struct {
	LogLevel  logx.LogLevel  `env:",omitempty"`
	LogFormat logx.LogFormat `env:",omitempty"`

	TraceCollectorEndpoint  string         `env:",omitempty"`
	MetricCollectorEndpoint string         `env:",omitempty"`
	MetricCollectInterval   types.Duration `env:",omitempty"`

	tracerProvider     *otelsdktracer.TracerProvider
	loggerProvider     *otelsdklogger.LoggerProvider
	metricProvider     *otelsdkmetric.MeterProvider
	prometheusGatherer promclient.Gatherer

	enabledLevel logx.LogLevel

	// registry log processor registry
	registry otelsdklogger.Processor
	// info *appinfo.Info `inject:",opt"`
}

func (o *Otel) SetDefault() {
	if o.MetricCollectorEndpoint != "" {
		if o.MetricCollectInterval == 0 {
			o.MetricCollectInterval = types.Duration(time.Minute)
		}
	}
}

func (o *Otel) Init(ctx context.Context) error {
	// TODO
	// if value, ok := appinfo.InfoFromContext(ctx); ok {
	// 	v.info = value
	// }
	appname, appversion := "todo", "version"

	o.registry = &registry{}

	// prometheus gatherer
	pg := promclient.NewRegistry()
	if err := pg.Register(promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{})); err != nil {
		return err
	}
	if err := pg.Register(promcollectors.NewGoCollector()); err != nil {
		return err
	}
	o.prometheusGatherer = pg

	pr, err := otelpromexporter.New(
		otelpromexporter.WithRegisterer(pg),
		otelpromexporter.WithoutScopeInfo(),
	)
	if err != nil {
		return err
	}

	tpopts := []otelsdktracer.TracerProviderOption{
		otelsdktracer.WithSampler(otelsdktracer.AlwaysSample()),
	}
	lpopts := []otelsdklogger.LoggerProviderOption{
		otelsdklogger.WithProcessor(otelsdklogger.NewSimpleProcessor(exporter.New(o.LogFormat))),
		otelsdklogger.WithProcessor(o.registry),
	}
	mpopts := []otelsdkmetric.Option{
		otelsdkmetric.WithReader(pr),
		metric.ViewsOption(),
	}

	res := otelsdkresource.NewSchemaless(
		semconv.ServiceName(appname),
		semconv.ServiceVersion(appversion),
	)
	tpopts = append(tpopts, otelsdktracer.WithResource(res))
	lpopts = append(lpopts, otelsdklogger.WithResource(res))
	mpopts = append(mpopts, otelsdkmetric.WithResource(res))

	if len(o.TraceCollectorEndpoint) > 0 {
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(o.TraceCollectorEndpoint),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithTimeout(3*time.Second),
		)

		exp, err := otlptrace.New(ctx, client)
		if err != nil {
			return err
		}
		tpopts = append(tpopts, otelsdktracer.WithBatcher(exporter.IgnoreErrSpanExporter(exp)))
	}

	if len(o.MetricCollectorEndpoint) > 0 {
		exp, err := otlpmetricgrpc.New(
			ctx,
			otlpmetricgrpc.WithEndpoint(o.MetricCollectorEndpoint),
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithTimeout(3*time.Second),
		)
		if err != nil {
			return err
		}
		reader := otelsdkmetric.NewPeriodicReader(
			exp,
			otelsdkmetric.WithInterval(time.Duration(o.MetricCollectInterval)),
		)

		mpopts = append(mpopts, otelsdkmetric.WithReader(reader))
	}

	o.loggerProvider = otelsdklogger.NewLoggerProvider(lpopts...)
	o.tracerProvider = otelsdktracer.NewTracerProvider(tpopts...)
	o.metricProvider = otelsdkmetric.NewMeterProvider(mpopts...)

	return nil
}

func (o *Otel) Close(c context.Context) error {
	eg, ctx := errgroup.WithContext(c)

	if tp := o.tracerProvider; tp != nil {
		eg.Go(func() error { return errors.Join(tp.ForceFlush(ctx), tp.Shutdown(ctx)) })
	}

	if lp := o.loggerProvider; lp != nil {
		eg.Go(func() error { return errors.Join(lp.ForceFlush(ctx), lp.Shutdown(ctx)) })
	}

	if mp := o.metricProvider; mp != nil {
		eg.Go(func() error { return errors.Join(mp.ForceFlush(ctx), mp.Shutdown(ctx)) })
	}

	return eg.Wait()
}

func (o *Otel) WithContext(ctx context.Context) context.Context {
	return contextx.Compose(
		providers.CarryTracerProvider(o.tracerProvider),
		providers.CarryLoggerProvider(o.loggerProvider),
		providers.CarryMetricProvider(o.metricProvider),
		providers.CarryMetricGatherer(o.prometheusGatherer),
		logx.Carry(logger.NewLogger(ctx, o.LogLevel)),
		CarryLogProcessorRegistry(o.registry.(LogProcessorRegistry)),
	)(ctx)
}
