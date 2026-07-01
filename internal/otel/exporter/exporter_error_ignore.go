package exporter

import (
	"context"

	otelsdktracer "go.opentelemetry.io/otel/sdk/trace"
)

func IgnoreErrSpanExporter(spanExporter otelsdktracer.SpanExporter) otelsdktracer.SpanExporter {
	return &errIgnoreExporter{
		SpanExporter: spanExporter,
	}
}

type errIgnoreExporter struct {
	otelsdktracer.SpanExporter
}

func (e *errIgnoreExporter) ExportSpans(ctx context.Context, span []otelsdktracer.ReadOnlySpan) error {
	_ = e.SpanExporter.ExportSpans(ctx, span)
	return nil
}
