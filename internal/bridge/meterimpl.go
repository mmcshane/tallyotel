package bridge

import (
	"context"
	"errors"
	"fmt"

	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
)

// ErrorUnsupportedInstrument is used to signal that an instrument cannot be
// created by a Meter because the instrument kind and number kind together are
// not supported.
var ErrorUnsupportedInstrument = errors.New("unsupported instrument")

type (
	// MeterImpl is an implementation of metric.MeterImpl that uses Tally and
	// wraps a tally.Scope
	MeterImpl struct {
		scope   tally.Scope
		buckets HistogramBucketer
	}

	syncScopeInstrument interface {
		// RecordOneInScope provides an optimized method or recording a value
		// when the scope is known a priori.
		RecordOneInScope(context.Context, tally.Scope, number.Number)
	}
)

// NewMeterImpl instantiates a MeterImpl wrapping the provided scope and using
// the provided bucket factory to configure buckets for histograms.
func NewMeterImpl(scope tally.Scope, buckets HistogramBucketer) *MeterImpl {
	return &MeterImpl{
		scope:   scope,
		buckets: buckets,
	}
}

// RecordBatch records multiple measurements to multiple instruments in a single
// call.
func (m *MeterImpl) RecordBatch(
	ctx context.Context,
	labels []attribute.KeyValue,
	measurements ...metric.Measurement,
) {
	scope := m.scope
	if len(labels) > 0 {
		scope = scope.Tagged(KVsToTags(labels))
	}
	for _, m := range measurements {
		ssi := m.SyncImpl().(syncScopeInstrument)
		ssi.RecordOneInScope(ctx, scope, m.Number())
	}
}

// NewSyncInstrument creates new SyncInstrument objects to support OTEL metric
// instruments. Supported instruments are Counter (int64 only), UpDownCounter
// (int64 only), Histogram. If a requested instrument is not supported the error
// returned here will satisfy errors.Is(err, ErrorUnsupportedInstrument).
func (m *MeterImpl) NewSyncInstrument(
	desc metric.Descriptor,
) (metric.SyncImpl, error) {
	switch desc.InstrumentKind() {
	case sdkapi.CounterInstrumentKind,
		sdkapi.UpDownCounterInstrumentKind:
		if desc.NumberKind() == number.Int64Kind {
			return NewCounter(desc, m.scope), nil
		}
	case sdkapi.HistogramInstrumentKind:
		return NewHistogram(desc, m.scope, m.buckets(desc)), nil
	}
	return nil, fmt.Errorf("%w: %v %v",
		ErrorUnsupportedInstrument, desc.InstrumentKind(), desc.NumberKind())
}

// NewAsyncInstrument is required by the metric.MeterImpl interface but no
// asynchronous instruments are supported because Tally doesn't support them.
func (m *MeterImpl) NewAsyncInstrument(
	desc metric.Descriptor,
	runner metric.AsyncRunner,
) (metric.AsyncImpl, error) {
	return nil, fmt.Errorf("%w: %v %v",
		ErrorUnsupportedInstrument, desc.InstrumentKind(), desc.NumberKind())
}
