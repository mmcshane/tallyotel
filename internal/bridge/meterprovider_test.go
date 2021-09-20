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
	scope := tally.NewTestScope("base", nil)
	mp := bridge.NewMeterProvider(scope, bridge.WithMeterScoper(
		func(name string, base tally.Scope) tally.Scope {
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

func TestSTandardMeterNaming(t *testing.T) {
	scope := tally.NewTestScope("base", nil)
	mp := bridge.NewMeterProvider(scope)
	m := metric.Must(mp.Meter("meter"))
	m.NewInt64Counter("c").Add(context.TODO(), 1)
	ctrSnaps := scope.Snapshot().Counters()

	_, ok := ctrSnaps["base.meter.c+"]
	require.True(t, ok, "counter name should be base.meter.c")
}
