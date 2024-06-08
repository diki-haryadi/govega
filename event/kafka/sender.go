package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/diki-haryadi/govega/event"
	"github.com/diki-haryadi/govega/log"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	KafkaSender struct {
		Brokers        []string               `json:"brokers" mapstructure:"brokers"`
		BatchSize      int                    `json:"batch_size" mapstructure:"batch_size"`
		BatchTimeout   string                 `json:"batch_timeout" mapstructure:"batch_timeout"`
		KeyFile        string                 `json:"key_file" mapstructure:"key_file"`
		CertFile       string                 `json:"cert_file" mapstructure:"cert_file"`
		CACertificate  string                 `json:"ca_cert" mapstructure:"ca_cert"`
		AuthType       string                 `json:"auth_type" mapstructure:"auth_type"`
		Username       string                 `json:"username" mapstructure:"username"`
		Password       string                 `json:"password" mapstructure:"password"`
		MaxAttempts    int                    `json:"max_attempts" mapstructure:"max_attempts"`
		Balancer       string                 `json:"balancer" mapstructure:"balancer"`
		BalancerConfig map[string]interface{} `json:"balancer_config" mapstructure:"balancer_config"`

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

		writer     *kafka.Writer
		propagator propagation.TextMapPropagator
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

	dialer, err := dial(kaf.CertFile, kaf.KeyFile, kaf.CACertificate, kaf.Username, kaf.Password, kaf.AuthType)
	if err != nil {
		return nil, err
	}

	balancer, err := getBalancer(kaf.Balancer, kaf.BalancerConfig)
	if err != nil {
		return nil, err
	}

	writerCfg := kafka.WriterConfig{
		Brokers:     kaf.Brokers,
		Balancer:    balancer,
		Dialer:      dialer,
		Logger:      newKafkaPrintLogger(kaf.PrintLogLevel),
		ErrorLogger: newKafkaErrorLogger(kaf.ErrorLogLevel),
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

	carrier := newKafkaMessageCarrier(&msg)
	k.propagator.Inject(ctx, carrier)

	err = k.writer.WriteMessages(ctx, msg)
	if err != nil {

		var writeErrors kafka.WriteErrors
		if errors.As(err, &writeErrors) {
			errs := make([]string, len(writeErrors))
			for i, writeErr := range writeErrors {
				errs[i] = writeErr.Error()
			}

			errMsg := strings.Join(errs, ", ")
			span.SetStatus(codes.Error, errMsg)

			return fmt.Errorf("%w: %s", err, errMsg)
		}

		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func newKafkaConsumeMessage(reader *kafka.Reader, message *KafkaMessageCarrier) *KafkaConsumeMessage {
	return &KafkaConsumeMessage{
		KafkaMessageCarrier: message,
		reader:              reader,
	}
}

func (k *KafkaConsumeMessage) GetEventConsumeMessage(ctx context.Context) (*event.EventConsumeMessage, error) {
	em, err := k.parseMessage(k.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	em.Topic = k.Topic
	if len(k.Key) > 0 {
		em.Key = string(k.Key)
	}

	return em, nil
}

func (k *KafkaConsumeMessage) Commit(ctx context.Context) error {
	if err := k.reader.CommitMessages(ctx, *k.Message); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	return nil
}

func (k *KafkaConsumeMessage) parseMessage(value []byte) (*event.EventConsumeMessage, error) {
	return event.NewEventConsumeMessage(value)
}

func getBalancer(balancer string, config map[string]interface{}) (kafka.Balancer, error) {
	switch balancer {
	case "hash":
		// TODO: provide way to use different hash mechanism
		// by default it will use FNV-1a
		return &kafka.Hash{}, nil
	case "round_robin":
		return &kafka.RoundRobin{}, nil
	case "crc32":
		consistent, ok := config["consistent"].(bool)
		if !ok {
			log.WithFields(log.Fields{
				"value":    config["consistent"],
				"balancer": balancer,
			}).Warn("invalid consistent, default to false")

			consistent = false
		}
		return &kafka.CRC32Balancer{
			Consistent: consistent,
		}, nil
	case "murmur2":
		consistent, ok := config["consistent"].(bool)
		if !ok {
			log.WithFields(log.Fields{
				"value":    config["consistent"],
				"balancer": balancer,
			}).Warn("invalid consistent, default to false")

			consistent = false
		}
		return &kafka.Murmur2Balancer{
			Consistent: consistent,
		}, nil
	case "least_bytes", "":
		return &kafka.LeastBytes{}, nil
	}

	return nil, fmt.Errorf("invalid balancer: %s", balancer)
}
