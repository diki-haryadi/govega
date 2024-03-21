package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/dikiharyadi19/govegapunk/event"
	"github.com/dikiharyadi19/govegapunk/log"
)

type (
	KafkaIterator struct {
		reader     *kafka.Reader
		tracer     trace.Tracer
		propagator propagation.TextMapPropagator
	}
)

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

	dialer, err := dial(
		kaf.CertFile,
		kaf.KeyFile,
		kaf.CACertificate,
		kaf.Username,
		kaf.Password,
		kaf.AuthType,
	)
	if err != nil {
		return nil, err
	}

	kafCfg := kafka.ReaderConfig{
		Brokers:               kaf.Brokers,
		Topic:                 topic,
		GroupID:               group,
		Dialer:                dialer,
		QueueCapacity:         kaf.QueueCapacity,
		MaxBytes:              kaf.MaxBytes,
		MinBytes:              kaf.MinBytes,
		WatchPartitionChanges: kaf.WatchPartitionChanges,
		StartOffset:           kaf.StartOffset,
	}

	if kaf.MaxWait != "" {
		maxWait, err := time.ParseDuration(kaf.MaxWait)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener max wait value: %w", err)
		}

		kafCfg.MaxWait = maxWait
	}

	if kaf.ReadlagInterval != "" {
		interval, err := time.ParseDuration(kaf.HeartbeatInterval)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener heartbeat interval value: %w", err)
		}

		kafCfg.HeartbeatInterval = interval
	}
	if kaf.CommitInterval != "" {
		interval, err := time.ParseDuration(kaf.CommitInterval)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener commit interval value: %w", err)
		}

		kafCfg.CommitInterval = interval
	}

	if kaf.PartitionWatchInterval != "" {
		interval, err := time.ParseDuration(kaf.PartitionWatchInterval)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener partition watch interval value: %w", err)
		}

		kafCfg.PartitionWatchInterval = interval
	}

	if kaf.SessionTimeout != "" {
		timeout, err := time.ParseDuration(kaf.SessionTimeout)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener session timeout value: %w", err)
		}

		kafCfg.SessionTimeout = timeout
	}

	if kaf.RebalanceTimeout != "" {
		timeout, err := time.ParseDuration(kaf.RebalanceTimeout)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener rebalance timeout value: %w", err)
		}

		kafCfg.RebalanceTimeout = timeout
	}

	if kaf.JoinGroupBackoff != "" {
		backoff, err := time.ParseDuration(kaf.JoinGroupBackoff)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener join group backoff value: %w", err)
		}

		kafCfg.JoinGroupBackoff = backoff
	}

	if kaf.RetentionTime != "" {
		retentionTime, err := time.ParseDuration(kaf.RetentionTime)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener retention time value: %w", err)
		}

		kafCfg.RetentionTime = retentionTime
	}

	if kaf.ReadBackoffMin != "" {
		min, err := time.ParseDuration(kaf.ReadBackoffMin)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener read backoff min value: %w", err)
		}

		kafCfg.ReadBackoffMin = min
	}

	if kaf.ReadBackoffMax != "" {
		max, err := time.ParseDuration(kaf.ReadBackoffMax)
		if err != nil {
			return nil, fmt.Errorf("[kafka/iterator] invalid kafka listener read backoff max value: %w", err)
		}

		kafCfg.ReadBackoffMax = max
	}

	if kaf.MaxAttempts > 0 {
		kafCfg.MaxAttempts = kaf.MaxAttempts
	}

	reader := kafka.NewReader(kafCfg)
	iter = &KafkaIterator{
		reader:     reader,
		tracer:     otel.Tracer("event/kafka"),
		propagator: otel.GetTextMapPropagator(),
	}

	err = nil
	return
}

func (k *KafkaIterator) Next(ctx context.Context) (event.ConsumeMessage, error) {
	msg, err := k.reader.FetchMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("[kafka/iterator] failed to read message: %w", err)
	}

	msgCarrier := newKafkaMessageCarrier(&msg)
	parentSpanContext := k.propagator.Extract(ctx, msgCarrier)

	consumeMessage := newKafkaConsumeMessage(k.reader, msgCarrier)
	_, span := k.injectOtel(parentSpanContext, msg, consumeMessage)
	defer span.End()

	return consumeMessage, nil
}

func (k *KafkaIterator) injectOtel(ctx context.Context, msg kafka.Message, msgCarrier propagation.TextMapCarrier) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKey.String("kafka"),
		semconv.MessagingProtocolKey.String("kafka"),
		semconv.MessagingDestinationKindKey.String("topic"),
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
	k.propagator.Inject(newctx, msgCarrier)
	return newctx, span
}

func (k *KafkaIterator) Close() error {
	return k.reader.Close()
}
