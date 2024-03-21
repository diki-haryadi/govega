package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/dikiharyadi19/govegapunk/event"
)

type (
	KafkaSender struct {
		Brokers       []string `json:"brokers" mapstructure:"brokers"`
		BatchSize     int      `json:"batch_size" mapstructure:"batch_size"`
		BatchTimeout  string   `json:"batch_timeout" mapstructure:"batch_timeout"`
		KeyFile       string   `json:"key_file" mapstructure:"key_file"`
		CertFile      string   `json:"cert_file" mapstructure:"cert_file"`
		CACertificate string   `json:"ca_cert" mapstructure:"ca_cert"`
		AuthType      string   `json:"auth_type" mapstructure:"auth_type"`
		Username      string   `json:"username" mapstructure:"username"`
		Password      string   `json:"password" mapstructure:"password"`
		MaxAttempts   int      `json:"max_attempts" mapstructure:"max_attempts"`
		writer        *kafka.Writer
		propagator    propagation.TextMapPropagator
	}
)

func NewKafkaSender(_ context.Context, config interface{}) (event.Sender, error) {
	var kaf KafkaSender
	if err := mapstructure.Decode(config, &kaf); err != nil {
		return nil, err
	}

	if len(kaf.Brokers) > 0 && strings.Contains(kaf.Brokers[0], ",") {
		bks := strings.Split(kaf.Brokers[0], ",")
		kaf.Brokers = bks
	}

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

	writerCfg := kafka.WriterConfig{
		Brokers:  kaf.Brokers,
		Balancer: &kafka.LeastBytes{},
		Dialer:   dialer,
	}

	if kaf.BatchSize != 0 {
		writerCfg.BatchSize = kaf.BatchSize
	}

	if kaf.BatchTimeout != "" {
		batchTimeout, err := time.ParseDuration(kaf.BatchTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka sender batch timeout value: %w", err)
		}
		writerCfg.BatchTimeout = batchTimeout
	}

	if kaf.MaxAttempts > 0 {
		writerCfg.MaxAttempts = kaf.MaxAttempts
	}

	pub := kafka.NewWriter(writerCfg)

	kaf.writer = pub
	kaf.propagator = otel.GetTextMapPropagator()
	return &kaf, nil
}

func (k *KafkaSender) Send(ctx context.Context, message *event.EventMessage) error {
	tr := otel.Tracer("event/emitter")
	ctx, span := tr.Start(ctx, "SEND "+message.Topic,
		trace.WithAttributes(semconv.MessagingOperationProcess),
		trace.WithAttributes(semconv.MessagingDestinationKindTopic),
		trace.WithAttributes(semconv.MessagingDestinationKey.String(message.Topic)),
		trace.WithAttributes(semconv.MessagingProtocolKey.String("kafka")),
		trace.WithAttributes(semconv.MessagingSystemKey.String("kafka")),
		trace.WithAttributes(semconv.MessagingConversationIDKey.String(message.Key)),
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	mb, err := message.ToBytes()
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: message.Topic,
		Value: mb,
	}

	if message.Key != "" {
		msg.Key = []byte(message.Key)
	}

	msgCarrier := newKafkaMessageCarrier(&msg)
	k.propagator.Inject(ctx, msgCarrier)

	if err = k.writer.WriteMessages(ctx, msg); err != nil {
		var writerErrors kafka.WriteErrors
		if errors.Is(err, &writerErrors) {
			errs := make([]string, len(writerErrors))
			for i, writeErr := range writerErrors {
				errs[i] = writeErr.Error()
			}

			errMsg := strings.Join(errs, ",")
			span.SetStatus(codes.Error, errMsg)

			return fmt.Errorf("%w: %s", err, errMsg)
		}

		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
