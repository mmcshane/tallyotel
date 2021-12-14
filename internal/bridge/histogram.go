package bridge

import (
	"context"
	"sync"
	"time"

	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
	"go.opentelemetry.io/otel/metric/unit"
)

type (
	histRecorder func(tally.Histogram, number.Number, number.Kind)

	// Histogram is a sdkapi.SyncImpl implementation that bridges between an
	// OTEL Histogram and a Tally Histogram.
	Histogram struct {
		desc      sdkapi.Descriptor
		baseScope tally.Scope
		record    histRecorder
		buckets   tally.Buckets

		initDefault sync.Once
		defaultHist tally.Histogram
	}
)

// NewHistogram instantiates a new Histogram that uses the proved scope and
// bucket configuration.
func NewHistogram(
	desc sdkapi.Descriptor,
	scope tally.Scope,
	buckets tally.Buckets,
) *Histogram {
	recorder := recordFloat64
	if desc.Unit() == unit.Milliseconds {
		recorder = recordMilliseconds
	}
	return &Histogram{
		desc:      desc,
		baseScope: scope,
		record:    recorder,
		buckets:   buckets,
	}
}

// Implementation is unused
func (h *Histogram) Implementation() interface{} {
	return nil
}

// Descriptor observes this Histogram's Descriptor object
func (h *Histogram) Descriptor() sdkapi.Descriptor {
	return h.desc
}

// RecordOne adds a value to this histogram.
func (h *Histogram) RecordOne(
	ctx context.Context,
	n number.Number,
	labels []attribute.KeyValue,
) {
	if len(labels) == 0 {
		h.recordToDefault(n)
		return
	}
	s := h.baseScope.Tagged(KVsToTags(labels))
	h.record(s.Histogram(h.desc.Name(), h.buckets), n, h.desc.NumberKind())
}

func (h *Histogram) recordToDefault(n number.Number) {
	h.initDefault.Do(func() {
		h.defaultHist = h.baseScope.Histogram(h.desc.Name(), h.buckets)
	})
	h.record(h.defaultHist, n, h.desc.NumberKind())
}

// RecordOneInScope is used to record a value when the scope can be provided by
// the caller. This is only known to be the case during Meter batch recordings.
func (h *Histogram) RecordOneInScope(
	ctx context.Context,
	scope tally.Scope,
	n number.Number,
) {
	if scope == h.baseScope {
		h.recordToDefault(n)
		return
	}
	h.record(scope.Histogram(h.desc.Name(), h.buckets), n, h.desc.NumberKind())
}

func recordFloat64(hist tally.Histogram, n number.Number, k number.Kind) {
	hist.RecordValue(n.CoerceToFloat64(k))
}

func recordMilliseconds(hist tally.Histogram, n number.Number, k number.Kind) {
	dur := time.Duration(n.CoerceToFloat64(k) * float64(time.Millisecond))
	hist.RecordDuration(dur)
}
