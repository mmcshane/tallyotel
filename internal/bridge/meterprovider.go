package bridge

import (
	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/registry"
)

type (
	// Opt is the type for optional arguments to a MeterProvider.
	Opt func(*MeterProvider)

	// HistogramBucketer maps metric metadata to a tally.Buckets instance. This
	// is necessary because OTEL does not have a way to indicate bucket
	// confguration at histogram creation time.
	HistogramBucketer func(metric.Descriptor) tally.Buckets

	// MeterProvider is an implementation of metric.Meterprovider wrapping a
	// tally.Scope.
	MeterProvider struct {
		scope   tally.Scope
		buckets HistogramBucketer
	}
)

// WithHistogramBucketer wraps a histogram bucket factory into a MeterProvider
// option
func WithHistogramBucketer(f HistogramBucketer) Opt {
	return func(mp *MeterProvider) {
		mp.buckets = f
	}
}

// NewMeterProvider creates a new MeterProvider wrapping the provided
// tally.Scope.
func NewMeterProvider(scope tally.Scope, opts ...Opt) metric.MeterProvider {
	mp := &MeterProvider{
		scope: scope,
		buckets: func(metric.Descriptor) tally.Buckets {
			return tally.DefaultBuckets
		},
	}
	for _, opt := range opts {
		opt(mp)
	}
	return mp
}

// Meter creates a new metric.Meter implementation that wraps a tally.Scope that
// is a sub-scope of the scope provided to this MeterProvider at construction
// time.
func (p *MeterProvider) Meter(
	instrumentationName string,
	opts ...metric.MeterOption,
) metric.Meter {
	impl := &MeterImpl{
		scope:   p.scope.SubScope(instrumentationName),
		buckets: p.buckets,
	}
	return metric.WrapMeterImpl(
		registry.NewUniqueInstrumentMeterImpl(impl),
		instrumentationName, opts...)
}
