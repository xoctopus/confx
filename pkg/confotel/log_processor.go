package confotel

import (
	"context"
	"sync"

	"github.com/xoctopus/x/contextx"
	otelsdklogger "go.opentelemetry.io/otel/sdk/log"
	"golang.org/x/sync/errgroup"
)

type LogProcessorRegistry interface {
	RegisterLogProcessor(p otelsdklogger.Processor)
}

type tCtxLogProcessorRegistry struct{}

var (
	LogProcessorRegistryFrom  = contextx.From[tCtxLogProcessorRegistry, LogProcessorRegistry]
	MustLogProcessorRegistry  = contextx.Must[tCtxLogProcessorRegistry, LogProcessorRegistry]
	WithLogProcessorRegistry  = contextx.With[tCtxLogProcessorRegistry, LogProcessorRegistry]
	CarryLogProcessorRegistry = contextx.Carry[tCtxLogProcessorRegistry, LogProcessorRegistry]
)

type registry struct {
	m sync.Map
}

var _ LogProcessorRegistry = (*registry)(nil)

func (r *registry) RegisterLogProcessor(p otelsdklogger.Processor) {
	r.m.Store(p, struct{}{})
}

func (r *registry) Enabled(ctx context.Context, param otelsdklogger.EnabledParameters) bool {
	return true
}

func (r *registry) OnEmit(ctx context.Context, record *otelsdklogger.Record) error {
	for k := range r.m.Range {
		_ = k.(otelsdklogger.Processor).OnEmit(ctx, record)
	}
	return nil
}

func (r *registry) Shutdown(ctx context.Context) error {
	g, c := errgroup.WithContext(ctx)

	for k := range r.m.Range {
		g.Go(func() error {
			return k.(otelsdklogger.Processor).Shutdown(c)
		})
	}

	return g.Wait()
}

func (r *registry) ForceFlush(ctx context.Context) error {
	g, c := errgroup.WithContext(ctx)

	for k := range r.m.Range {
		g.Go(func() error {
			return k.(otelsdklogger.Processor).ForceFlush(c)
		})
	}

	return g.Wait()
}
