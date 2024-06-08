// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	//"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Filter returns true if the request should be traced.
type Filter func(*http.Request) bool

// SpanNameFormatter creates a custom span name from the operation and request object.
type SpanNameFormatter func(operation string, req *http.Request) string

// config provides configuration for the beego OpenTelemetry
// middleware. Configuration is modified using the provided Options.
type config struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	propagators    propagation.TextMapPropagator
	filters        []Filter
	formatter      SpanNameFormatter
}

// Option applies a configuration to the given config.
type Option interface {
	apply(*config)
}

// optionFunc is a function type that applies a particular
// configuration to the beego middleware in question.
type optionFunc func(c *config)

// Apply will apply the option to the config, c.
func (o optionFunc) apply(c *config) {
	o(c)
}

// ------------------------------------------ Options

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *config) {
		cfg.tracerProvider = provider
	})
}

// WithMeterProvider specifies a meter provider to use for creating a meter.
// If none is specified, the global provider is used.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return optionFunc(func(cfg *config) {
		cfg.meterProvider = provider
	})
}

// WithPropagators sets the propagators used in the middleware.
// Defaults to global.Propagators().
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(c *config) {
		c.propagators = propagators
	})
}

// WithFilter adds the given filter for use in the middleware.
// Defaults to no filters.
func WithFilter(f Filter) Option {
	return optionFunc(func(c *config) {
		c.filters = append(c.filters, f)
	})
}

// WithSpanNameFormatter sets the formatter to be used to format
// span names. Defaults to the path template.
func WithSpanNameFormatter(f SpanNameFormatter) Option {
	return optionFunc(func(c *config) {
		c.formatter = f
	})
}

// ------------------------------------------ Private Functions

func newConfig(options ...Option) *config {
	config := &config{
		tracerProvider: otel.GetTracerProvider(),
		meterProvider:  otel.GetMeterProvider(),
		propagators:    propagation.NewCompositeTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)), propagation.TraceContext{}, propagation.Baggage{}),
		filters:        []Filter{},
		formatter:      defaultSpanNameFormatter,
	}
	for _, option := range options {
		option.apply(config)
	}
	return config
}

// defaultSpanNameFormatter is the default formatter for spans created with the beego
// integration. Returns the route path template, or the URL path if the current path
// is not associated with a router.
func defaultSpanNameFormatter(operation string, req *http.Request) string {
	return req.Method + " " + req.URL.Path
}
