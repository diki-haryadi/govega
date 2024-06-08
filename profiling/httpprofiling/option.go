package httpprofiling

import (
	"time"

	"github.com/diki-haryadi/govega/router"
)

// WithReadTimeout, set http profiler timeout to read the request.
//
// Default: 5 minutes
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.ReadTimeout = timeout
	}
}

// WithReadTimeout, set http profiler timeout to write the response.
//
// Default: 5 minutes
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.WriteTimeout = timeout
	}
}

// WithPort, set the http profiler port
//
// Default: 8432
func WithPort(port int) Option {
	return func(c *config) {
		c.Port = port
	}
}

// WithRouter, set the default router which is used by the http profiler
func WithRouter(r *router.MyRouter) Option {
	return func(c *config) {
		c.Router = r
	}
}

// WithManualStart, indicate whether or not to automatically start the server on init
// If set to true, the caller need to manually start the server by calling `Start` function
//
// Default: false
func WithManualStart(manualStart bool) Option {
	return func(c *config) {
		c.ManualStart = manualStart
	}
}
