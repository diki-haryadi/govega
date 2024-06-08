package monitor

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/diki-haryadi/govega/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpLatencyHistogram      *prometheus.HistogramVec
	httpResponsesTotalCounter *prometheus.CounterVec
	httpMetricLabels          = []string{"handler", "method", "httpcode", "env"}

	grpcLatencyHistogram      *prometheus.HistogramVec
	grpcResponsesTotalCounter *prometheus.CounterVec
	grpcMetricLabels          = []string{"handler", "grpccode", "grpcstatus", "env"}

	consumerLatencyHistogram      *prometheus.HistogramVec
	consumerResponsesTotalCounter *prometheus.CounterVec
	consumerMetricLabels          = []string{"topic", "group", "status", "env"}
)

func init() {
	Init("")
}

func Init(appName string) {
	// metrics name doesn't accept dash(-) character
	appName = strings.ReplaceAll(appName, "-", "_")

	registerHistogram(appName)
	registerCounter(appName)
}

func registerHistogram(appName string) {
	unregister(httpLatencyHistogram)
	httpLatencyHistogram = createAndRegisterHistogram("http", appName, httpMetricLabels)

	unregister(grpcLatencyHistogram)
	grpcLatencyHistogram = createAndRegisterHistogram("grpc", appName, grpcMetricLabels)

	unregister(consumerLatencyHistogram)
	consumerLatencyHistogram = createAndRegisterHistogram("consumer", appName, consumerMetricLabels)
}

func registerCounter(appName string) {
	unregister(httpResponsesTotalCounter)
	httpResponsesTotalCounter = createAndRegisterCounter("http", appName, httpMetricLabels)

	unregister(grpcResponsesTotalCounter)
	grpcResponsesTotalCounter = createAndRegisterCounter("grpc", appName, grpcMetricLabels)

	unregister(consumerResponsesTotalCounter)
	consumerResponsesTotalCounter = createAndRegisterCounter("consumer", appName, consumerMetricLabels)
}

func unregister(c prometheus.Collector) {
	if !reflect.ValueOf(c).IsNil() {
		prometheus.Unregister(c)
	}
}

func createAndRegisterHistogram(metric, namespace string, labels []string) *prometheus.HistogramVec {
	newHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      fmt.Sprintf("%s_duration_seconds", metric),
		Namespace: namespace,
		Help:      fmt.Sprintf("the latency of %s calls", metric),
	}, labels)
	if err := prometheus.Register(newHistogram); err != nil {
		log.WithFields(log.Fields{
			"metric":    metric,
			"namespace": namespace,
		}).WithError(err).Warnln("[monitor] unable to register histogram")
	}

	return newHistogram
}

func createAndRegisterCounter(metric, namespace string, labels []string) *prometheus.CounterVec {
	newCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      fmt.Sprintf("%s_responses_total", metric),
		Namespace: namespace,
		Help:      fmt.Sprintf("The count of %s responses issued", metric),
	}, labels)
	if err := prometheus.Register(newCounter); err != nil {
		log.WithFields(log.Fields{
			"metric":    metric,
			"namespace": namespace,
		}).WithError(err).Warnln("[monitor] unable to register counter")
	}

	return newCounter
}
