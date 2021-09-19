package bridge_test

import (
	"context"
	"testing"
	"time"

	"github.com/mmcshane/tallyotel/internal/bridge"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
	"go.opentelemetry.io/otel/metric/unit"
)

func TestF64Histogram(t *testing.T) {
	scope := tally.NewTestScope("scope", nil)
	hist := bridge.NewHistogram(
		metric.NewDescriptor(
			"h",
			sdkapi.HistogramInstrumentKind,
			number.Float64Kind),
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

func TestDurationHistogram(t *testing.T) {
	scope := tally.NewTestScope("scope", nil)
	hist := bridge.NewHistogram(
		metric.NewDescriptor(
			"h",
			sdkapi.HistogramInstrumentKind,
			number.Float64Kind,
			metric.WithUnit(unit.Milliseconds)),
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

func TestBoundHistogram(t *testing.T) {
	scope := tally.NewTestScope("scope", nil)
	hist := bridge.NewHistogram(
		metric.NewDescriptor(
			"h",
			sdkapi.HistogramInstrumentKind,
			number.Float64Kind),
		scope,
		tally.MustMakeLinearValueBuckets(0.0, 1.0, 5))

	kvs := []attribute.KeyValue{attribute.Key("foo").String("bar")}

	// these two lines should record to the same histogram
	hist.Bind(kvs).RecordOne(context.TODO(), number.NewFloat64Number(1.5))
	hist.RecordOne(context.TODO(), number.NewFloat64Number(2.1), kvs)

	_, ok := scope.Snapshot().Histograms()["scope.h+"]
	require.False(t, ok, "Never recorded to the hist without labels so "+
		"unlabeled hist should not exist")

	snap, ok := scope.Snapshot().Histograms()["scope.h+foo=bar"]
	require.True(t, ok)
	vals := snap.Values()
	require.EqualValues(t, 0, vals[1.0])
	require.EqualValues(t, 1, vals[2.0])
	require.EqualValues(t, 1, vals[3.0])
	require.EqualValues(t, 0, vals[4.0])
}
