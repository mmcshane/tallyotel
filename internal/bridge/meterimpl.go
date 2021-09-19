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

type (
	MeterImpl struct {
		scope   tally.Scope
		buckets HistogramBucketer
	}

	SyncScopeInstrument interface {
		RecordOneInScope(context.Context, tally.Scope, number.Number)
	}
)

var UnsupportedInstrumentError = errors.New("unsupported instrument")

func NewMeterImpl(scope tally.Scope, buckets HistogramBucketer) *MeterImpl {
	return &MeterImpl{
		scope:   scope,
		buckets: buckets,
	}
}

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
		ssi := m.SyncImpl().(SyncScopeInstrument)
		ssi.RecordOneInScope(ctx, scope, m.Number())
	}
}

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
		UnsupportedInstrumentError, desc.InstrumentKind(), desc.NumberKind())
}

func (m *MeterImpl) NewAsyncInstrument(
	desc metric.Descriptor,
	runner metric.AsyncRunner,
) (metric.AsyncImpl, error) {
	return nil, fmt.Errorf("%w: %v %v",
		UnsupportedInstrumentError, desc.InstrumentKind(), desc.NumberKind())
}
