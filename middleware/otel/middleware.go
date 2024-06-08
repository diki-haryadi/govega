package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Handler implements the http.Handler interface and provides
// trace and metrics to beego web apps.
type Handler struct {
	http.Handler
}

// ServerHTTP calls the configured handler to serve HTTP for req to rr.
func (o *Handler) ServeHTTP(rr http.ResponseWriter, req *http.Request) {
	o.Handler.ServeHTTP(rr, req)
}

// NewMiddleWare creates a MiddleWare that provides OpenTelemetry
// Parameter service should describe the name of the (virtual) server handling the request.
// The OTelMiddleWare can be configured using the provided Options.
func NewMiddleWare(service string, options ...Option) func(http.Handler) http.Handler {
	cfg := newConfig(options...)

	httpOptions := []otelhttp.Option{
		otelhttp.WithTracerProvider(cfg.tracerProvider),
		otelhttp.WithMeterProvider(cfg.meterProvider),
		otelhttp.WithPropagators(cfg.propagators),
	}

	for _, f := range cfg.filters {
		httpOptions = append(
			httpOptions,
			otelhttp.WithFilter(otelhttp.Filter(f)),
		)
	}

	if cfg.formatter != nil {
		httpOptions = append(httpOptions, otelhttp.WithSpanNameFormatter(cfg.formatter))
	}

	return func(handler http.Handler) http.Handler {
		return &Handler{
			otelhttp.NewHandler(
				handler,
				service,
				httpOptions...,
			),
		}
	}
}
