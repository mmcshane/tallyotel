package bridge_test

import (
	"context"
	"testing"

	"github.com/mmcshane/tallyotel/internal/bridge"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
)

func testCounter(
	scope string,
	ctr string,
	kind sdkapi.InstrumentKind,
) (tally.TestScope, *bridge.Counter) {
	tscope := tally.NewTestScope(scope, nil)
	bctr := bridge.NewCounter(
		metric.NewDescriptor(ctr, kind, number.Int64Kind), tscope)
	return tscope, bctr
}

func TestIncrOnlyCounter(t *testing.T) {
	scope, ctr := testCounter("scope", "ctr", sdkapi.CounterInstrumentKind)

	require.Panics(t, func() {
		ctr.RecordOne(context.TODO(), number.NewInt64Number(-1), nil)
	}, "otel counter instruments are increment-only")

	ctr.RecordOne(context.TODO(), number.NewInt64Number(1), nil)

	snap, ok := scope.Snapshot().Counters()["scope.ctr+"]
	require.True(t, ok)
	require.EqualValues(t, 1, snap.Value())
}

func TestTaggedRecord(t *testing.T) {
	scope, ctr := testCounter("scope", "ctr", sdkapi.CounterInstrumentKind)

	ctr.RecordOne(context.TODO(), number.NewInt64Number(1),
		[]attribute.KeyValue{attribute.Key("foo").Int(1)})

	_, ok := scope.Snapshot().Counters()["scope.ctr+"]
	require.False(t, ok, "scope.ctr should not be registered in scope")

	snap, ok := scope.Snapshot().Counters()["scope.ctr+foo=1"]
	require.True(t, ok)
	require.EqualValues(t, 1, snap.Value())
}

func TestBoundCounter(t *testing.T) {
	scope, unbound := testCounter("scope", "ctr", sdkapi.CounterInstrumentKind)
	ctr := unbound.Bind([]attribute.KeyValue{attribute.Key("foo").String("bar")})

	require.Panics(t, func() {
		ctr.RecordOne(context.TODO(), number.NewInt64Number(-1))
	}, "otel counter instruments are increment-only")

	ctr.RecordOne(context.TODO(), number.NewInt64Number(1))

	_, ok := scope.Snapshot().Counters()["scope.ctr+"]
	require.False(t, ok, "scope.ctr should not be registered in scope")

	snap, ok := scope.Snapshot().Counters()["scope.ctr+foo=bar"]
	require.True(t, ok)
	require.EqualValues(t, 1, snap.Value())
}

func TestUpDownCounter(t *testing.T) {
	scope, ctr := testCounter("scope", "ctr", sdkapi.UpDownCounterInstrumentKind)

	ctr.RecordOne(context.TODO(), number.NewInt64Number(3), nil)
	ctr.RecordOne(context.TODO(), number.NewInt64Number(-5), nil)

	snap, ok := scope.Snapshot().Counters()["scope.ctr+"]
	require.True(t, ok)
	require.EqualValues(t, 3-5, snap.Value())
}
