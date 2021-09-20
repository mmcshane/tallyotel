package bridge

import (
	"context"
	"sync"
	"time"

	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/unit"
)

type (
	histRecorder func(tally.Histogram, number.Number, number.Kind)

	// Histogram is a metric.SyncImpl implementation that bridges between an
	// OTEL Histogram and a Tally Histogram.
	Histogram struct {
		desc      metric.Descriptor
		baseScope tally.Scope
		record    histRecorder
		buckets   tally.Buckets

		initDefault sync.Once
		defaultHist tally.Histogram
	}

	// BoundHistogram is a Histogram that has been bound to a set of
	// attribute.KeyValue labels.
	BoundHistogram struct {
		desc   metric.Descriptor
		hist   tally.Histogram
		record histRecorder
	}
)

// NewHistogram instantiates a new Histogram that uses the proved scope and
// bucket configuration.
func NewHistogram(
	desc metric.Descriptor,
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
func (h *Histogram) Descriptor() metric.Descriptor {
	return h.desc
}

// Bind transforms this Histogram into a BoundHistogram, embedding the provided
// labels. A new scope is created from the base scope initially provided at
// construction time.
func (h *Histogram) Bind(labels []attribute.KeyValue) metric.BoundSyncImpl {
	newScope := h.baseScope.Tagged(KVsToTags(labels))
	return &BoundHistogram{
		desc:   h.desc,
		hist:   newScope.Histogram(h.desc.Name(), h.buckets),
		record: h.record,
	}
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

// RecordOne adds a value to this BoundHistogram.
func (h *BoundHistogram) RecordOne(ctx context.Context, n number.Number) {
	h.record(h.hist, n, h.desc.NumberKind())
}

// Unbind is a no op. It's not clear what it is supposed to do.
func (h *BoundHistogram) Unbind() {}

func recordFloat64(hist tally.Histogram, n number.Number, k number.Kind) {
	hist.RecordValue(n.CoerceToFloat64(k))
}

func recordMilliseconds(hist tally.Histogram, n number.Number, k number.Kind) {
	dur := time.Duration(n.CoerceToFloat64(k) * float64(time.Millisecond))
	hist.RecordDuration(dur)
}
