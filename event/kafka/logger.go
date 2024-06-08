package kafka

import (
	"github.com/diki-haryadi/govega/log"
)

var (
	defaultPrintLogLevel = "debug"
	defaultErrorLogLevel = "error"
)

// newKafkaPrintLogger, create new instance of kafka print logger
func newKafkaPrintLogger(level string) kafka.Logger {
	if level == "" {
		level = defaultPrintLogLevel
	}

	return newKafkaLogger(level)
}

// newKafkaErrorLogger create new instance of kafka error logger
func newKafkaErrorLogger(level string) kafka.Logger {
	if level == "" {
		level = defaultErrorLogLevel
	}

	return newKafkaLogger(level)
}

func newKafkaLogger(level string) kafka.Logger {
	switch level {
	case "panic", "fatal":
		return kafka.LoggerFunc(log.Fatalf)
	case "error":
		return kafka.LoggerFunc(log.Errorf)
	case "warning":
		return kafka.LoggerFunc(log.Warnf)
	case "info":
		return kafka.LoggerFunc(log.Infof)
	case "debug":
		return kafka.LoggerFunc(log.Debugf)
	case "discard":
		return kafka.LoggerFunc(discardf)
	default:
		return kafka.LoggerFunc(log.Infof)
	}
}

func discardf(_ string, _ ...interface{}) {
	// do nothing, discard log
}
