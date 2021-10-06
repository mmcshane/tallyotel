package bridge_test

import (
	"context"
	"testing"

	"github.com/mmcshane/tallyotel/internal/bridge"
	"github.com/stretchr/testify/require"
	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/metric"
)

func TestMeterReuseProviderScope(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("base", nil)
	mp := bridge.NewMeterProvider(scope, bridge.WithMeterScoper(
		func(_ []string, base tally.Scope) tally.Scope {
			return base
		}))
	m := metric.Must(mp.Meter("meter"))
	m.NewInt64Counter("c").Add(context.TODO(), 1)
	ctrSnaps := scope.Snapshot().Counters()

	_, ok := ctrSnaps["base.meter.c+"]
	require.False(t, ok, "meter should reuse parent provider named scope")

	_, ok = ctrSnaps["base.c+"]
	require.True(t, ok, "counter name should be base.c")
}

func TestSTandardMeterNamingDoubleScope(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("", nil)
	mp := bridge.NewMeterProvider(scope, bridge.WithMeterScoper(
		func(_ []string, base tally.Scope) tally.Scope {
			return base.SubScope("x").SubScope("y")
		}))
	m := metric.Must(mp.Meter("meter"))
	m.NewInt64Counter("c").Add(context.TODO(), 1)
	ctrSnaps := scope.Snapshot().Counters()

	_, ok := ctrSnaps["x.y.c+"]
	require.True(t, ok, "counter name should be x.y.c")
}

func TestStandardMeterNaming(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("base", nil)
	mp := bridge.NewMeterProvider(scope)
	m := metric.Must(mp.Meter("meter"))
	m.NewInt64Counter("c").Add(context.TODO(), 1)
	ctrSnaps := scope.Snapshot().Counters()

	_, ok := ctrSnaps["base.meter.c+"]
	require.True(t, ok, "counter name should be base.meter.c")
}

func TestMeterNameSplitting(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("base", nil)
	mp := bridge.NewMeterProvider(scope, bridge.WithScopeNameSeparator("@"))
	m := metric.Must(mp.Meter("@foo@bar@baz@@@"))
	m.NewInt64Counter("c").Add(context.TODO(), 1)
	ctrSnaps := scope.Snapshot().Counters()

	_, ok := ctrSnaps["base.foo.bar.baz.c+"]
	require.True(t, ok, "counter name should be base.foo.bar.baz.c")
}
