package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/diki-haryadi/govega/event"
	"github.com/diki-haryadi/govega/log"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	KafkaListener struct {
		Brokers       []string `json:"brokers" mapstructure:"brokers"`
		KeyFile       string   `json:"key_file" mapstructure:"key_file"`
		CertFile      string   `json:"cert_file" mapstructure:"cert_file"`
		CACertificate string   `json:"ca_cert" mapstructure:"ca_cert"`
		AuthType      string   `json:"auth_type" mapstructure:"auth_type"`
		Username      string   `json:"username" mapstructure:"username"`
		Password      string   `json:"password" mapstructure:"password"`

		// PrintLogLevel will set default log level for all log message other than error message
		// available level are: panic, fatal, error, warning, info, debug, discard
		// discard level will discard any of the log message received
		// default: debug
		PrintLogLevel string `json:"print_log_level" mapstructure:"print_log_level"`
		// ErrorLogLevel will set default log level for error message log
		// available level are: panic, fatal, error, warning, info, debug, discard
		// discard level will discard any of the log message received
		// default: error
		ErrorLogLevel string `json:"error_log_level" mapstructure:"error_log_level"`

		QueueCapacity          int    `json:"queue_capacity" mapstructure:"queue_capacity"`
		MinBytes               int    `json:"min_bytes" mapstructure:"min_bytes"`
		MaxBytes               int    `json:"max_bytes" mapstructure:"max_bytes"`
		MaxWait                string `json:"max_wait" mapstructure:"max_wait"`
		MaxAttempts            int    `json:"max_attempts" mapstructure:"max_attempts"`
		ReadlagInterval        string `json:"read_lag_interval" mapstructure:"read_lag_interval"`
		HeartbeatInterval      string `json:"heartbeat_interval" mapstructure:"heartbeat_interval"`
		CommitInterval         string `json:"commit_interval" mapstructure:"commit_interval"`
		PartitionWatchInterval string `json:"partition_watch_interval" mapstructure:"partition_watch_interval"`
		WatchPartitionChanges  bool   `json:"watch_partition_changes" mapstructure:"watch_partition_changes"`
		SessionTimeout         string `json:"session_timeout" mapstructure:"session_timeout"`
		RebalanceTimeout       string `json:"rebalance_timeout" mapstructure:"rebalance_timeout"`
		JoinGroupBackoff       string `json:"join_group_backoff" mapstructure:"join_group_backoff"`
		RetentionTime          string `json:"retention_time" mapstructure:"retention_time"`
		StartOffset            int64  `json:"start_offset" mapstructure:"start_offset"`
		ReadBackoffMin         string `json:"read_backoff_min" mapstructure:"read_backoff_min"`
		ReadBackoffMax         string `json:"read_backoff_max" mapstructure:"read_backoff_max"`
	}

	KafkaIterator struct {
		reader     *kafka.Reader
		tracer     trace.Tracer
		propagator propagation.TextMapPropagator
	}

	KafkaConsumeMessage struct {
		*KafkaMessageCarrier
		reader *kafka.Reader
	}
)

func NewKafkaListener(_ context.Context, config interface{}) (event.Listener, error) {
	var listener KafkaListener
	if err := mapstructure.Decode(config, &listener); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &listener, nil
}

func (k *KafkaListener) Listen(ctx context.Context, topic, group string) (event.Iterator, error) {
	return newKafkaIterator(k, topic, group)
}

func newKafkaIterator(kaf *KafkaListener, topic, group string) (iter *KafkaIterator, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{
				"panic": r,
			}).Errorln("panic on creating kafka iterator")
			iter = nil
			err = fmt.Errorf("failed to create kafka iterator: %v", r)
			return
		}
	}()

	dialer, err := dial(kaf.CertFile, kaf.KeyFile, kaf.CACertificate, kaf.Username, kaf.Password, kaf.AuthType)
	if err != nil {
		return nil, err
	}

	// NOTE: NewReader may panic
	conf := kafka.ReaderConfig{
		Brokers:               kaf.Brokers,
		Topic:                 topic,
		GroupID:               group,
		Dialer:                dialer,
		QueueCapacity:         kaf.QueueCapacity,
		MaxBytes:              kaf.MaxBytes,
		MinBytes:              kaf.MinBytes,
		WatchPartitionChanges: kaf.WatchPartitionChanges,
		StartOffset:           kaf.StartOffset,
		Logger:                newKafkaPrintLogger(kaf.PrintLogLevel),
		ErrorLogger:           newKafkaErrorLogger(kaf.ErrorLogLevel),
	}

	if kaf.MaxWait != "" {
		maxWait, err := time.ParseDuration(kaf.MaxWait)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener max wait value: %w", err)
		}

		conf.MaxWait = maxWait
	}

	if kaf.ReadlagInterval != "" {
		interval, err := time.ParseDuration(kaf.ReadlagInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener read lag interval value: %w", err)
		}

		conf.ReadLagInterval = interval
	}

	if kaf.HeartbeatInterval != "" {
		interval, err := time.ParseDuration(kaf.HeartbeatInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener heartbeat interval value: %w", err)
		}

		conf.HeartbeatInterval = interval
	}

	if kaf.CommitInterval != "" {
		interval, err := time.ParseDuration(kaf.CommitInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener commit interval value: %w", err)
		}

		conf.CommitInterval = interval
	}

	if kaf.PartitionWatchInterval != "" {
		interval, err := time.ParseDuration(kaf.PartitionWatchInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener partition watch interval value: %w", err)
		}

		conf.PartitionWatchInterval = interval
	}

	if kaf.SessionTimeout != "" {
		timeout, err := time.ParseDuration(kaf.SessionTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener session timeout value: %w", err)
		}

		conf.SessionTimeout = timeout
	}

	if kaf.RebalanceTimeout != "" {
		timeout, err := time.ParseDuration(kaf.RebalanceTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener rebalance timeout value: %w", err)
		}

		conf.RebalanceTimeout = timeout
	}

	if kaf.JoinGroupBackoff != "" {
		backoff, err := time.ParseDuration(kaf.JoinGroupBackoff)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener join group backoff value: %w", err)
		}

		conf.JoinGroupBackoff = backoff
	}

	if kaf.RetentionTime != "" {
		retentionTime, err := time.ParseDuration(kaf.RetentionTime)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener retention time value: %w", err)
		}

		conf.RetentionTime = retentionTime
	}

	if kaf.ReadBackoffMin != "" {
		min, err := time.ParseDuration(kaf.ReadBackoffMin)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener read backoff min value: %w", err)
		}

		conf.ReadBackoffMin = min
	}

	if kaf.ReadBackoffMax != "" {
		max, err := time.ParseDuration(kaf.ReadBackoffMax)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka listener read backoff max value: %w", err)
		}

		conf.ReadBackoffMax = max
	}

	if kaf.MaxAttempts > 0 {
		conf.MaxAttempts = kaf.MaxAttempts
	}

	reader := kafka.NewReader(conf)

	iter = &KafkaIterator{
		reader:     reader,
		tracer:     otel.Tracer("event/consumer"),
		propagator: otel.GetTextMapPropagator(),
	}
	err = nil
	return
}

func (k *KafkaIterator) Next(ctx context.Context) (event.ConsumeMessage, error) {
	msg, err := k.reader.FetchMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	carrier := newKafkaMessageCarrier(&msg)
	parentSpanContext := k.propagator.Extract(ctx, carrier)

	consumeMessage := newKafkaConsumeMessage(k.reader, carrier)
	_, span := k.injectOtel(parentSpanContext, msg, consumeMessage)
	defer span.End()

	return consumeMessage, nil
}

func (k *KafkaIterator) injectOtel(ctx context.Context, msg kafka.Message,
	carrier propagation.TextMapCarrier) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKey.String("kafka"),
		semconv.MessagingProtocolKey.String("kafka"),
		semconv.MessagingDestinationKindTopic,
		semconv.MessagingDestinationKey.String(msg.Topic),
		semconv.MessagingOperationReceive,
		semconv.MessagingMessageIDKey.String(strconv.FormatInt(msg.Offset, 10)),
		semconv.MessagingConversationIDKey.String(string(msg.Key)),
		KafkaPartitionKey.Int64(int64(msg.Partition)),
		KafkaConsumerGroupKey.String(event.GetConsumerGroupFromContext(ctx)),
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	newctx, span := k.tracer.Start(ctx, fmt.Sprintf("kafka.consume.%s", msg.Topic), opts...)
	k.propagator.Inject(newctx, carrier)
	return newctx, span
}

func (k *KafkaIterator) Close() error {
	return k.reader.Close()
}
