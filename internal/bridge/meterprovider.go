package bridge

import (
	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/registry"
)

type (
	Opt func(*MeterProvider)

	HistogramBucketer func(metric.Descriptor) tally.Buckets

	MeterProvider struct {
		scope   tally.Scope
		buckets HistogramBucketer
	}
)

func WithHistogramBucketer(f HistogramBucketer) Opt {
	return func(mp *MeterProvider) {
		mp.buckets = f
	}
}

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
