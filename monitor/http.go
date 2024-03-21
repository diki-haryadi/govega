package monitor

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/superman-tech/lib/env"
)

// FeedHTTPMetrics to monitor latency, http code counts
func FeedHTTPMetrics(status int, duration time.Duration, path string, method string) {
	// TODO: find out if we really need this "all" handler or is it safe to be remove?
	httpLatencyHistogram.With(prometheus.Labels{
		"handler": "all",
		"method":  method, "httpcode": fmt.Sprintf("%d", status),
		"env": env.Get(),
	}).Observe(duration.Seconds())

	httpLatencyHistogram.With(prometheus.Labels{
		"handler":  path,
		"method":   method,
		"httpcode": fmt.Sprintf("%d", status),
		"env":      env.Get(),
	}).Observe(duration.Seconds())

	// TODO: find out if we really need this "all" handler or is it safe to be remove?
	httpResponsesTotalCounter.With(prometheus.Labels{
		"handler":  "all",
		"method":   method,
		"httpcode": fmt.Sprintf("%d", status),
		"env":      env.Get(),
	}).Inc()

	httpResponsesTotalCounter.With(prometheus.Labels{
		"handler":  path,
		"method":   method,
		"httpcode": fmt.Sprintf("%d", status),
		"env":      env.Get(),
	}).Inc()
}
