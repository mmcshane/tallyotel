package bridge

import (
	"context"
	"sync"
	"time"

	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/unit"
)

type (
	histRecorder func(tally.Histogram, number.Number, number.Kind)

	Histogram struct {
		desc      metric.Descriptor
		baseScope tally.Scope
		record    histRecorder
		buckets   tally.Buckets

		initDefault sync.Once
		defaultHist tally.Histogram
	}

	BoundHistogram struct {
		desc   metric.Descriptor
		hist   tally.Histogram
		record histRecorder
	}
)

func NewHistogram(
	desc metric.Descriptor,
	scope tally.Scope,
	buckets tally.Buckets,
) *Histogram {
	recorder := RecordFloat64
	if desc.Unit() == unit.Milliseconds {
		recorder = RecordMilliseconds
	}
	return &Histogram{
		desc:      desc,
		baseScope: scope,
		record:    recorder,
		buckets:   buckets,
	}
}

func (h *Histogram) Implementation() interface{} {
	return nil
}

func (h *Histogram) Descriptor() metric.Descriptor {
	return h.desc
}

func (h *Histogram) Bind(labels []attribute.KeyValue) metric.BoundSyncImpl {
	newScope := h.baseScope.Tagged(KVsToTags(labels))
	return &BoundHistogram{
		desc:   h.desc,
		hist:   newScope.Histogram(h.desc.Name(), h.buckets),
		record: h.record,
	}
}

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

func (h *BoundHistogram) RecordOne(ctx context.Context, n number.Number) {
	h.record(h.hist, n, h.desc.NumberKind())
}

func (c *BoundHistogram) Unbind() {
	panic("not implemented")
}

func RecordFloat64(hist tally.Histogram, n number.Number, k number.Kind) {
	hist.RecordValue(n.CoerceToFloat64(k))
}

func RecordMilliseconds(hist tally.Histogram, n number.Number, k number.Kind) {
	dur := time.Duration(n.CoerceToFloat64(k) * float64(time.Millisecond))
	hist.RecordDuration(dur)
}
