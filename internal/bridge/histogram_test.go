package bridge_test

import (
	"context"
	"testing"
	"time"

	"github.com/mmcshane/tallyotel/internal/bridge"
	"github.com/stretchr/testify/require"
	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
	"go.opentelemetry.io/otel/metric/unit"
)

/*
func TestTallyIsBusted(t *testing.T) {
	scope := tally.NewTestScope("scope", nil)
	h := scope.Histogram("foo", nil)

	h.RecordValue(0.33)

	snap, ok := scope.Snapshot().Histograms()["scope.foo+"]
	require.True(t, ok)
	vals := snap.Values()
	require.Greater(t, len(vals), 0, "we recorded a value, where did it go?")
}
*/

func TestInt64HistToFloat(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("scope", nil)
	mp := bridge.NewMeterProvider(scope)
	m := metric.Must(mp.Meter("m"))
	hist := m.NewInt64Histogram("h1")

	hist.Record(context.TODO(), 3)

	snap, ok := scope.Snapshot().Histograms()["scope.m.h1+"]
	require.True(t, ok)
	vals := snap.Values()
	require.EqualValues(t, 1, vals[5.0])
}

func TestF64Histogram(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("scope", nil)
	hist := bridge.NewHistogram(
		sdkapi.NewDescriptor(
			"h",
			sdkapi.HistogramInstrumentKind,
			number.Float64Kind,
			"description",
			unit.Dimensionless),
		scope,
		tally.MustMakeLinearValueBuckets(0.0, 1.0, 5))

	hist.RecordOne(context.TODO(), number.NewFloat64Number(0.5), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(1.5), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(2.5), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(3.5), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(3.5), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(3.5), nil)

	snap, ok := scope.Snapshot().Histograms()["scope.h+"]
	require.True(t, ok)
	vals := snap.Values()
	require.EqualValues(t, 1, vals[1.0])
	require.EqualValues(t, 1, vals[2.0])
	require.EqualValues(t, 1, vals[3.0])
	require.EqualValues(t, 3, vals[4.0])
}

func TestInt64Histogram(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("scope", nil)
	hist := bridge.NewHistogram(
		sdkapi.NewDescriptor(
			"h",
			sdkapi.HistogramInstrumentKind,
			number.Int64Kind,
			"description",
			unit.Dimensionless),
		scope,
		tally.MustMakeLinearValueBuckets(0.5, 1.0, 5))

	hist.RecordOne(context.TODO(), number.NewInt64Number(1), nil)
	hist.RecordOne(context.TODO(), number.NewInt64Number(2), nil)
	hist.RecordOne(context.TODO(), number.NewInt64Number(3), nil)
	hist.RecordOne(context.TODO(), number.NewInt64Number(3), nil)

	snap, ok := scope.Snapshot().Histograms()["scope.h+"]
	require.True(t, ok)
	vals := snap.Values()
	require.EqualValues(t, 0, vals[0.5])
	require.EqualValues(t, 1, vals[1.5])
	require.EqualValues(t, 1, vals[2.5])
	require.EqualValues(t, 2, vals[3.5])
}

func TestDurationHistogram(t *testing.T) {
	t.Parallel()
	scope := tally.NewTestScope("scope", nil)
	hist := bridge.NewHistogram(
		sdkapi.NewDescriptor(
			"h",
			sdkapi.HistogramInstrumentKind,
			number.Float64Kind,
			"description",
			unit.Milliseconds),
		scope,
		tally.MustMakeLinearDurationBuckets(0, 1*time.Second, 5))

	hist.RecordOne(context.TODO(), number.NewFloat64Number(500), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(1500), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(2500), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(3500), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(3500), nil)
	hist.RecordOne(context.TODO(), number.NewFloat64Number(3500), nil)

	snap, ok := scope.Snapshot().Histograms()["scope.h+"]
	require.True(t, ok)
	vals := snap.Durations()
	require.EqualValues(t, 1, vals[1*time.Second])
	require.EqualValues(t, 1, vals[2*time.Second])
	require.EqualValues(t, 1, vals[3*time.Second])
	require.EqualValues(t, 3, vals[4*time.Second])
}
