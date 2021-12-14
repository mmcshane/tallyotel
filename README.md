# tallyotel

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mmcshane/tallyotel)](https://pkg.go.dev/github.com/mmcshane/tallyotel)
[![Go Report Card](https://goreportcard.com/badge/github.com/mmcshane/tallyotel)](https://goreportcard.com/report/github.com/mmcshane/tallyotel)

A [Tally](https://github.com/uber-go/tally)/[Open
Telemetry](https://github.com/open-telemetry/opentelemetry-go) bridge allowing
for code using Open Telemetry instruments to emit metrics with Tally.

A demonstration of emitting metrics from Open Telemetry instruments through
Tally to Prometheus can be found at https://github.com/mmcshane/tallyotel-demo.

As Tally contains abstractions for exporting metrics and Open Telemetry _also_
contains abstractions for exporting metrics, the use of this particular bridge
library is likely to be a transient period for any given codebase. If your code
is instrumented to use Open Telemetry already then you're probably better off
migrating to the Open Telemetry native exporter configuration. However, if
you're in a situation where the use of Tally is required but you still want to
use Open Telemetry metric instruments, then this library is what you need to
bridge the two.

## Instrument Support and Mappings

As Tally does not support asynchronous instrument types (i.e. the Observer
types), this library does not offer them. Here we describe how Open Telemetry
instrument types are mapped to Tally instruments.

| OTEL Type                  | Tally Type        | Notes                            |
|----------------------------|-------------------|----------------------------------|
| Counter                    | `tally.Counter`   | Only integer counters (`number.NumberKindInt64`) are supported. Note that OTEL counters are monotonic - use `UpDownCounter` to both increment and decrement. |
| Asynchronous Counter       |                   | Async instruments not supported. |
| Asynchronous Gauge         |                   | Async instruments not supported. |
| Histogram                  | `tally.Histogram` | Histograms using a unit of `unit.Millisecond` use the Tally `Histogram.RecordDuration` Histogram API, otherwise `Histogram.RecordValue`. |
| UpDownCounter              | `tally.Counter`   | Only integer counters (`number.NumberKindInt64`) are supported. |
| Asynchronous UpDownCounter |                   | Async instruments not supported. |

## Tally Scope Usage

Tally makes heavy use of a graph of scope objects for instrument naming and to
bundle "tags" with a set of instruments. Open Telemetry does not have a clear
analog for Tally scopes but this bridge does use them internally in a
predictable way. Here are the rules for how scopes are used within `tallyotel`.

1. Every `tallyotel.MeterProvider` instance has a Scope which is passed to it at
   construction time. Users can use this scope to make Tally-specific
   configurations to the metrics supplied by the `metric.Meter`(s) derived from
   a given top-level `tallyotel.MeterProvider`
1. Every `metric.Meter` created by a `tallyotel.MeterProvider` uses the provided
   Meter name to build a set of nested sub-scopes (see `tally.Scope.SubScope`)
   of the parent `tallyotel.MeterProvider`'s scope. Sub-scopes are implied
   through the Meter name via a separator string (by default: `"."`). The exact
   behavior here can be modified through a client-provided
   `tallyotel.MeterScoper`.
1. Instruments created by a `metric.Meter` and invoked without any
   `attribute.KeyValue`s use their parent `metric.Meter`'s scope
1. Instruments invoked with `attribute.KeyValue`s use a scope that is a
   sub-scope of their parent Meter's scope, tagged with the appropriate
   key-values (see `tally.Scope.Tagged`)

