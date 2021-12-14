package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"

	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
)

// ErrNonMonotonicValue is a base error cause returned when a negative value is
// added to an increase-only Counter.
var ErrNonMonotonicValue = errors.New("unexpected non-monotonic value")

type (
	// Counter implements the sdkapi.SyncImpl interface wrapping a tally.Counter
	Counter struct {
		desc      sdkapi.Descriptor
		baseScope tally.Scope

		initDefault sync.Once
		defaultCtr  tally.Counter
	}
)

// NewCounter instantiates a new Counter that uses the provided scope as its
// base scope.
func NewCounter(desc sdkapi.Descriptor, scope tally.Scope) *Counter {
	return &Counter{desc: desc, baseScope: scope}
}

// Implementation is unused
func (c *Counter) Implementation() interface{} {
	return nil
}

// Descriptor observes this Counter's Descriptor object
func (c *Counter) Descriptor() sdkapi.Descriptor {
	return c.desc
}

// RecordOne increments this counter by the provided value. If this Counter is
// configured to be an UpDownCounter then negative values are allowed.
func (c *Counter) RecordOne(
	ctx context.Context,
	n number.Number,
	labels []attribute.KeyValue,
) {
	value := n.AsInt64()
	if err := validateInt64(c.desc.InstrumentKind(), value); err != nil {
		otel.Handle(err)
		return
	}
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

// RecordOneInScope is used to record a value when the scope can be provided by
// the caller. This is only known to be the case during Meter batch recordings.
func (c *Counter) RecordOneInScope(
	ctx context.Context,
	scope tally.Scope,
	n number.Number,
) {
	value := n.AsInt64()
	if err := validateInt64(c.desc.InstrumentKind(), n.AsInt64()); err != nil {
		otel.Handle(err)
		return
	}
	if scope == c.baseScope {
		c.recordValidValueToDefault(value)
		return
	}
	scope.Counter(c.desc.Name()).Inc(value)
}

func validateInt64(kind sdkapi.InstrumentKind, value int64) error {
	if kind.Monotonic() && value < 0 {
		return fmt.Errorf("%w: %v", ErrNonMonotonicValue, value)
	}
	return nil
}
