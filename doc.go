// Package tallyotel provides a bridge between code instrumented to emit metrics
// via the the standard Open Telemetry go library/sdk
// (https://github.com/open-telemetry/opentelemetry-go) and the Tally
// (https://github.com/uber-go/tally) metrics library. Effectively it allows
// you to use Open Telmetry to _record_ your metrics and Tally to _emit_ your
// metrics.
package tallyotel
