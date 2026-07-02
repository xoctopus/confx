package exporter

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/go-json-experiment/json/jsontext"
	"go.opentelemetry.io/otel/log"
	otelsdkloggerp "go.opentelemetry.io/otel/sdk/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/xoctopus/confx/internal/otel/consts"
)

func NewWriter(r otelsdkloggerp.Record, format consts.Format) *Writer {
	w := &Writer{
		format: format,
		output: os.Stdout,
		record: r,
	}

	if r.Severity() >= log.SeverityWarn {
		w.output = os.Stderr
	}

	return w
}

type Writer struct {
	format consts.Format
	output io.Writer
	record otelsdkloggerp.Record
}

func (w *Writer) Write() error {
	switch w.format {
	case consts.TEXT:
		return w.WriteText()
	default:
		return w.WriteJSON()
	}
}

func (w *Writer) severity() string {
	if txt := w.record.SeverityText(); txt != "" {
		return txt
	}

	s := w.record.Severity()
	if s >= log.SeverityFatal {
		return "err"
	}
	if s >= log.SeverityError {
		return "err"
	}
	if s >= log.SeverityWarn {
		return "wrn"
	}
	if s >= log.SeverityInfo {
		return "inf"
	}
	if s >= log.SeverityDebug {
		return "deb"
	}
	return "deb"
}

func (w *Writer) WriteJSON() error {
	b := bytes.NewBuffer(nil)
	enc := jsontext.NewEncoder(b)

	tokens := []jsontext.Token{
		jsontext.BeginObject,
		jsontext.String(consts.KEY_LOG_TIMESTAMP), jsontext.String(w.record.Timestamp().Format(consts.TIME_FORMAT)),
		jsontext.String(consts.KEY_LOG_LEVEL), jsontext.String(w.severity()),
		jsontext.String(consts.KEY_TRACE_ID), jsontext.String(w.record.TraceID().String()),
		jsontext.String(consts.KEY_TRACE_SPAN_ID), jsontext.String(w.record.SpanID().String()),
		jsontext.String(consts.KEY_TRACE_SPAN_NAME), jsontext.String(w.record.InstrumentationScope().Name),
	}

	for _, attr := range w.record.Resource().Attributes() {
		key := ""
		switch attr.Key {
		case semconv.ServiceNameKey:
			key = consts.KEY_SERVICE_NAME
		case semconv.ServiceVersionKey:
			key = consts.KEY_SERVICE_VERSION
		default:
			continue
		}
		tokens = append(tokens, jsontext.String(key))
		tokens = append(tokens, TokenFor(attr.Value.AsInterface())...)
	}

	heads := make([]jsontext.Token, 0, 1)
	tails := make([]jsontext.Token, 0, w.record.AttributesLen()-1)

	seen := map[string]struct{}{}
	for attr := range w.record.WalkAttributes {
		if _, ok := seen[attr.Key]; ok {
			continue
		}
		if strings.HasPrefix(attr.Key, "@") {
			heads = append(heads, jsontext.String(attr.Key))
			heads = append(heads, TokenFor(LogValue(attr.Value))...)
		} else {
			tails = append(tails, jsontext.String(attr.Key))
			tails = append(tails, TokenFor(LogValue(attr.Value))...)
		}
		seen[attr.Key] = struct{}{}
	}

	tokens = append(tokens, append(heads, tails...)...)
	tokens = append(
		tokens,
		jsontext.String(consts.KEY_LOG_MESSAGE), jsontext.String(w.record.Body().AsString()),
		jsontext.EndObject,
	)

	for _, t := range tokens {
		if err := enc.WriteToken(t); err != nil {
			return err
		}
	}

	_, err := io.Copy(w.output, b)

	return err
}

func (w *Writer) WriteText() error {
	prefix := color.CyanString("%s:", w.record.SpanID().String())

	level := strings.ToUpper(w.severity())
	switch level {
	case "DEB":
		level = color.BlueString("DEB")
	case "INF":
		level = color.GreenString("INF")
	case "WRN":
		level = color.YellowString("WRN")
	case "ERR":
		level = color.RedString("ERR")
	}

	_, _ = fmt.Fprint(w.output, prefix)
	_, _ = fmt.Fprint(w.output, " ")
	_, _ = fmt.Fprint(w.output, level)
	_, _ = fmt.Fprint(w.output, " ")
	_, _ = fmt.Fprint(w.output, color.WhiteString(w.record.Timestamp().Format(consts.TIME_FORMAT)))
	_, _ = fmt.Fprint(w.output, " ")
	_, _ = fmt.Fprint(w.output, w.record.Body().AsString())

	attrs := map[string]any{}
	if name := w.record.InstrumentationScope().Name; name != "" {
		attrs[consts.KEY_TRACE_SPAN_NAME] = name
	}

	seen := map[string]struct{}{}
	for attr := range w.record.WalkAttributes {
		if _, ok := seen[attr.Key]; ok {
			continue
		}
		seen[attr.Key] = struct{}{}
		attrs[attr.Key] = LogValue(attr.Value)
	}

	keys := slices.Collect(maps.Keys(attrs))
	sort.Strings(keys)

	_, _ = fmt.Fprint(w.output, "\n")
	for _, k := range keys {
		_, _ = fmt.Fprintf(w.output, "\t%s: %v\n", k, LogAnyValue(attrs[k]))
	}
	return nil
}
