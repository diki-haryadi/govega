package monitor

import (
	"fmt"
	"time"

	"github.com/diki-haryadi/govega/env"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
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

// FeedGRPCMetrics to monitor grpc latency, status code counts
func FeedGRPCMetrics(method string, code codes.Code, duration time.Duration) {
	grpcLatencyHistogram.With(prometheus.Labels{
		"handler":    method,
		"grpccode":   fmt.Sprintf("%d", code),
		"grpcstatus": code.String(),
		"env":        env.Get(),
	}).Observe(duration.Seconds())

	grpcResponsesTotalCounter.With(prometheus.Labels{
		"handler":    method,
		"grpccode":   fmt.Sprintf("%d", code),
		"grpcstatus": code.String(),
		"env":        env.Get(),
	}).Inc()
}

// FeedConsumerMetrics to monitor consumer latency, status counts
func FeedConsumerMetrics(topic, group, status string, duration time.Duration) {
	consumerLatencyHistogram.With(prometheus.Labels{
		"topic":  topic,
		"group":  group,
		"status": status,
		"env":    env.Get(),
	}).Observe(duration.Seconds())

	consumerResponsesTotalCounter.With(prometheus.Labels{
		"topic":  topic,
		"group":  group,
		"status": status,
		"env":    env.Get(),
	}).Inc()
}
