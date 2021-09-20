package tallyotel

import (
	"github.com/mmcshane/tallyotel/internal/bridge"
	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/metric"
)

type (
	// Opt is the type for supplying optional configuration parameters to a
	// `tallyotel.NewMeterProvider`
	Opt = bridge.Opt

	// HistogramBucketer is an func allowing client code to pick different
	// bucketization strategies for histograms based on the information in the
	// histogram's metric.Descriptor.
	HistogramBucketer = bridge.HistogramBucketer
)

var (
	// WithHistogramBucketer wraps a HistogramBucketer into a tallyotel Opt so
	// that it can be passed in to a MeterProvider.
	WithHistogramBucketer = bridge.WithHistogramBucketer

	// DefaultBucketer returns the default histogram buckets. It is exposed here
	// for use as a fallback bucketing strategy within a custom
	// HistogramBucketer.
	DefaultBucketer = bridge.DefaultBucketer
)

// NewMeterProvider instantiates a tallyotel bridge metric.MeterProvider that
// uses the supplied tally.Scope as a base scope for the creation of child
// Meters and Instruments.
func NewMeterProvider(scope tally.Scope, opts ...Opt) metric.MeterProvider {
	return bridge.NewMeterProvider(scope, opts...)
}
