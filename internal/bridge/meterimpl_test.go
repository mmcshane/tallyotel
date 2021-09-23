package bridge_test

import (
	"context"
	"testing"

	"github.com/mmcshane/tallyotel/internal/bridge"
	"github.com/stretchr/testify/require"
	tally "github.com/uber-go/tally/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
)

func key(name string, kvs []attribute.KeyValue) string {
	return tally.KeyForPrefixedStringMap(name, bridge.KVsToTags(kvs))
}

func buckets(metric.Descriptor) tally.Buckets {
	return tally.MustMakeLinearValueBuckets(0.0, 1.0, 5)
}

func TestBatchRecord(t *testing.T) {
	for _, tt := range [...]struct {
		name   string
		labels []attribute.KeyValue
	}{
		{
			name: "no labels",
		},
		{
			name:   "foo=bar",
			labels: []attribute.KeyValue{attribute.Key("foo").String("bar")},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scope := tally.NewTestScope("scope", nil)
			prov := bridge.NewMeterProvider(
				scope, bridge.WithHistogramBucketer(buckets))
			meter := prov.Meter("meter")
			ctr := metric.Must(meter).NewInt64Counter("a")
			hist := metric.Must(meter).NewFloat64Histogram("h")

			var batchErrs []error
			withOTELErrorHandler(captureInto(&batchErrs), func() {
				meter.RecordBatch(context.TODO(), tt.labels,
					ctr.Measurement(4),
					ctr.Measurement(-1), // non-monotonic, should error
					hist.Measurement(1.5))
			})

			require.Len(t, batchErrs, 1)
			require.ErrorIs(t, batchErrs[0], bridge.ErrorNonMonotonicValue)

			snap := scope.Snapshot()

			ctrsnap, ok := snap.Counters()[key("scope.meter.a", tt.labels)]
			require.True(t, ok)
			require.EqualValues(t, 4, ctrsnap.Value())

			histsnap, ok := snap.Histograms()[key("scope.meter.h", tt.labels)]
			require.True(t, ok)
			require.EqualValues(t, 1, histsnap.Values()[2.0])
		})
	}
}

func TestUnsupported(t *testing.T) {
	m := bridge.NewMeterImpl(tally.NewTestScope("scope", nil), buckets)
	_, err := m.NewAsyncInstrument(metric.NewDescriptor(
		"name",
		sdkapi.CounterObserverInstrumentKind,
		number.Int64Kind,
	), nil)
	// none of the async instruments are supported
	require.ErrorIs(t, err, bridge.ErrorUnsupportedInstrument)

	_, err = m.NewSyncInstrument(metric.NewDescriptor(
		"name",
		sdkapi.CounterInstrumentKind,
		number.Float64Kind, // only integer histograms are supported
	))
	require.ErrorIs(t, err, bridge.ErrorUnsupportedInstrument)
}
