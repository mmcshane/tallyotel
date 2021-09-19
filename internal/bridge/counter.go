package bridge

import (
	"context"
	"sync"

	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
)

type (
	Counter struct {
		desc      metric.Descriptor
		baseScope tally.Scope

		initDefault sync.Once
		defaultCtr  tally.Counter
	}

	BoundCounter struct {
		desc metric.Descriptor
		ctr  tally.Counter
	}
)

func NewCounter(desc metric.Descriptor, scope tally.Scope) *Counter {
	return &Counter{desc: desc, baseScope: scope}
}

func (c *Counter) Implementation() interface{} {
	return nil
}

func (c *Counter) Descriptor() metric.Descriptor {
	return c.desc
}

func (c *Counter) Bind(labels []attribute.KeyValue) metric.BoundSyncImpl {
	newScope := c.baseScope.Tagged(KVsToTags(labels))
	return &BoundCounter{
		desc: c.desc,
		ctr:  newScope.Counter(c.desc.Name()),
	}
}

func (c *Counter) RecordOne(
	ctx context.Context,
	n number.Number,
	labels []attribute.KeyValue,
) {
	value := n.AsInt64()
	validateInt64(c.desc.InstrumentKind(), value)
	if len(labels) == 0 {
		c.recordValidValueToDefault(value)
		return
	}
	scope := c.baseScope.Tagged(KVsToTags(labels))
	scope.Counter(c.desc.Name()).Inc(value)
}

func (c *Counter) recordValidValueToDefault(valid int64) {
	c.initDefault.Do(func() {
		c.defaultCtr = c.baseScope.Counter(c.desc.Name())
	})
	c.defaultCtr.Inc(valid)
}

func (c *Counter) RecordOneInScope(
	ctx context.Context,
	scope tally.Scope,
	n number.Number,
) {
	value := n.AsInt64()
	validateInt64(c.desc.InstrumentKind(), value)
	if scope == c.baseScope {
		c.recordValidValueToDefault(value)
		return
	}
	scope.Counter(c.desc.Name()).Inc(value)
}

func (c *BoundCounter) RecordOne(ctx context.Context, n number.Number) {
	validateInt64(c.desc.InstrumentKind(), n.AsInt64())
	c.ctr.Inc(int64(n))
}

func (c *BoundCounter) Unbind() {
	panic("not implemented")
}

func validateInt64(kind sdkapi.InstrumentKind, value int64) {
	if kind.Monotonic() && value < 0 {
		panic("Monotonic counter incremented by a negative number")
	}
}
