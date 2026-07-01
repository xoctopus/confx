package exporter

import (
	"context"

	otelsdkloggerp "go.opentelemetry.io/otel/sdk/log"

	"github.com/xoctopus/confx/internal/otel/consts"
)

func New(format consts.Format) otelsdkloggerp.Exporter {
	return &exporter{format: format}
}

type exporter struct {
	format consts.Format
}

func (e *exporter) Export(_ context.Context, records []otelsdkloggerp.Record) error {
	for _, r := range records {
		w := NewWriter(r, e.format)
		if err := w.Write(); err != nil {
			return err
		}
	}

	return nil
}

func (e *exporter) ForceFlush(_ context.Context) error { return nil }

func (e *exporter) Shutdown(_ context.Context) error { return nil }
