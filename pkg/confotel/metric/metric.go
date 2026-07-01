package metric

import (
	"fmt"
	"time"

	"github.com/xoctopus/x/syncx"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Metric struct {
	Name        string
	Unit        string
	Description string
	Views       []View
}

type View struct {
	Instrument otelsdkmetric.Instrument
	Stream     otelsdkmetric.Stream
}

var views = syncx.NewSmap[string, []View]()

func ViewsOption() otelsdkmetric.Option {
	_views := make([]otelsdkmetric.View, 0)

	for _, vv := range views.Range {
		for _, v := range vv {
			_views = append(_views, otelsdkmetric.NewView(v.Instrument, v.Stream))
		}
	}

	return otelsdkmetric.WithView(_views...)
}

func NewOption(name string, appliers ...OptionFunc) *option {
	o := &option{}
	o.Name = name

	for i := range appliers {
		appliers[i](o)
	}

	// records views
	if len(o.Views) > 0 {
		views.Store(name, o.Views)
	}

	return o
}

type option struct {
	Metric
}

type OptionFunc = func(o *option)

func WithUnit(unit string) OptionFunc {
	return func(o *option) {
		o.Unit = unit
	}
}

func WithDescription(description string) OptionFunc {
	return func(o *option) {
		o.Description = description
	}
}

func WithView(view func(m Metric) View) OptionFunc {
	return func(o *option) {
		o.Views = append(o.Views, view(o.Metric))
	}
}

func WithAggregation(aggregation otelsdkmetric.Aggregation) OptionFunc {
	return WithView(func(m Metric) View {
		return View{
			Instrument: otelsdkmetric.Instrument{
				Name:        m.Name,
				Unit:        m.Unit,
				Description: m.Description,
			},
			Stream: otelsdkmetric.Stream{
				Name:        m.Name,
				Aggregation: aggregation,
			},
		}
	})
}

func WithAggregationFunc(typ string, d time.Duration) OptionFunc {
	return WithView(func(m Metric) View {
		return View{
			Instrument: otelsdkmetric.Instrument{
				Kind: otelsdkmetric.InstrumentKindObservableGauge,
				Name: fmt.Sprintf("%s__%s.%0.0fs", m.Name, typ, d.Seconds()),
				Unit: m.Unit,
			},
		}
	})
}
