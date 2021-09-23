package bridge_test

import "go.opentelemetry.io/otel"

var (
	panicHandler  otel.ErrorHandlerFunc = func(err error) { panic(err.Error()) }
	ignoreHandler otel.ErrorHandlerFunc = func(err error) {}
)

func init() {
	otel.SetErrorHandler(ignoreHandler)
}

func captureInto(errs *[]error) otel.ErrorHandlerFunc {
	return func(err error) {
		*errs = append(*errs, err)
	}
}

func withOTELErrorHandler(h otel.ErrorHandler, f func()) {
	orig := otel.GetErrorHandler()
	otel.SetErrorHandler(h)
	defer otel.SetErrorHandler(orig)
	f()
}
