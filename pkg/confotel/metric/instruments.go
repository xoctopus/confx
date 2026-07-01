package metric

import (
	"context"

	otelapimetric "go.opentelemetry.io/otel/metric"

	"github.com/xoctopus/confx/internal/otel/providers"
)

type Int64Counter interface {
	Add(ctx context.Context, incr int64, options ...otelapimetric.AddOption)
}

type Int64Recorder interface {
	Record(ctx context.Context, incr int64, options ...otelapimetric.RecordOption)
}

type Float64Counter interface {
	Add(ctx context.Context, incr float64, options ...otelapimetric.AddOption)
}

type Float64Recorder interface {
	Record(ctx context.Context, incr float64, options ...otelapimetric.RecordOption)
}

func NewInt64Counter(name string, appliers ...OptionFunc) Int64Counter {
	o := NewOption(name, appliers...)

	return &int64Instrument{
		option: o,
		counter: func(meter otelapimetric.Meter) (Int64Counter, error) {
			return meter.Int64Counter(o.Name, otelapimetric.WithUnit(o.Unit), otelapimetric.WithDescription(o.Description))
		},
	}
}

func NewInt64Histogram(name string, appliers ...OptionFunc) Int64Recorder {
	o := NewOption(name, appliers...)

	return &int64Instrument{
		option: o,
		histogram: func(meter otelapimetric.Meter) (Int64Recorder, error) {
			return meter.Int64Histogram(o.Name, otelapimetric.WithUnit(o.Unit), otelapimetric.WithDescription(o.Description))
		},
	}
}

type int64Instrument struct {
	*option
	counter   func(meter otelapimetric.Meter) (Int64Counter, error)
	histogram func(meter otelapimetric.Meter) (Int64Recorder, error)
}

func (i *int64Instrument) Add(ctx context.Context, incr int64, options ...otelapimetric.AddOption) {
	if c, err := i.counter(providers.MeterFrom(ctx)); err == nil {
		c.Add(ctx, incr, options...)
	}
}

func (i *int64Instrument) Record(ctx context.Context, incr int64, options ...otelapimetric.RecordOption) {
	if c, err := i.histogram(providers.MeterFrom(ctx)); err == nil {
		c.Record(ctx, incr, options...)
	}
}

func NewFloat64UpDownCounter(name string, appliers ...OptionFunc) Float64Counter {
	o := NewOption(name, appliers...)

	return &float64Instrument{
		option: o,
		counter: func(meter otelapimetric.Meter) (Float64Counter, error) {
			return meter.Float64UpDownCounter(o.Name, otelapimetric.WithUnit(o.Unit), otelapimetric.WithDescription(o.Description))
		},
	}
}

func NewFloat64Counter(name string, appliers ...OptionFunc) Float64Counter {
	o := NewOption(name, appliers...)

	return &float64Instrument{
		option: o,
		counter: func(meter otelapimetric.Meter) (Float64Counter, error) {
			return meter.Float64Counter(o.Name, otelapimetric.WithUnit(o.Unit), otelapimetric.WithDescription(o.Description))
		},
	}
}

func NewFloat64Histogram(name string, appliers ...OptionFunc) Float64Recorder {
	o := NewOption(name, appliers...)

	return &float64Instrument{
		option: o,
		histogram: func(meter otelapimetric.Meter) (Float64Recorder, error) {
			return meter.Float64Histogram(o.Name, otelapimetric.WithUnit(o.Unit), otelapimetric.WithDescription(o.Description))
		},
	}
}

type float64Instrument struct {
	*option
	counter   func(meter otelapimetric.Meter) (Float64Counter, error)
	histogram func(meter otelapimetric.Meter) (Float64Recorder, error)
}

func (i *float64Instrument) Add(ctx context.Context, incr float64, options ...otelapimetric.AddOption) {
	if c, err := i.counter(providers.MeterFrom(ctx)); err == nil {
		c.Add(ctx, incr, options...)
	}
}

func (i *float64Instrument) Record(ctx context.Context, incr float64, options ...otelapimetric.RecordOption) {
	if c, err := i.histogram(providers.MeterFrom(ctx)); err == nil {
		c.Record(ctx, incr, options...)
	}
}
