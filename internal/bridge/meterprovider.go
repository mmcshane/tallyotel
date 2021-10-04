package bridge

import (
	"strings"
	"time"

	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/unit"
)

var (

	// These are copied from the tally codebase. There's a bug in Tally where if
	// you don't supply ValueBuckets then Tally thinks that it _must_ be a
	// duration histogram and will _reject_ calls to Histogram.RecordValue. Thus
	// in order have a function value histogram, we _must_ pass a
	// tally.ValueBuckets. There's a commented-out test at the top of
	// histogram_test.go that illustrates the bug.

	defaultDurationBuckets = tally.DurationBuckets{
		0 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
		50 * time.Millisecond,
		75 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
		600 * time.Millisecond,
		800 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
	}

	defaultValueBuckets = tally.ValueBuckets(defaultDurationBuckets.AsValues())
)

type (
	// Opt is the type for optional arguments to a MeterProvider.
	Opt func(*MeterProvider)

	// HistogramBucketer maps metric metadata to a tally.Buckets instance. This
	// is necessary because OTEL does not have a way to indicate bucket
	// confguration at histogram creation time.
	HistogramBucketer func(metric.Descriptor) tally.Buckets

	// MeterScoper is a factory for scopes to be used in a Meter given a meter
	// name and a base scope.
	MeterScoper func(nameParts []string, baseScope tally.Scope) tally.Scope

	// MeterProvider is an implementation of metric.Meterprovider wrapping a
	// tally.Scope.
	MeterProvider struct {
		scope       tally.Scope
		buckets     HistogramBucketer
		meterScoper MeterScoper
		separator   string
	}
)

func defaultMeterScoper(nameParts []string, base tally.Scope) tally.Scope {
	scope := base
	for _, name := range nameParts {
		scope = scope.SubScope(name)
	}
	return scope
}

// WithScopeNameSeparator provides a string to a MeterProvider at construction
// time to be used in splitting child Meter names into scope names.
func WithScopeNameSeparator(s string) Opt {
	return func(mp *MeterProvider) {
		mp.separator = s
	}
}

// WithMeterScoper provides a MeterScoper to a MeterProvider at construction
// time
func WithMeterScoper(f MeterScoper) Opt {
	return func(mp *MeterProvider) {
		mp.meterScoper = f
	}
}

// DefaultBucketer is a HistogramBucketer that gives a hardcoded set of default
// buckets.
func DefaultBucketer(desc metric.Descriptor) tally.Buckets {
	if desc.Unit() == unit.Milliseconds {
		return append(tally.DurationBuckets(nil), defaultDurationBuckets...)
	}
	return append(tally.ValueBuckets(nil), defaultValueBuckets...)
}

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
		scope:       scope,
		buckets:     DefaultBucketer,
		meterScoper: defaultMeterScoper,
		separator:   tally.DefaultSeparator,
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
	trimmed := strings.Trim(instrumentationName, p.separator)
	parts := strings.Split(trimmed, p.separator)
	impl := &MeterImpl{
		scope:   p.meterScoper(parts, p.scope),
		buckets: p.buckets,
	}
	return metric.WrapMeterImpl(impl)
}
