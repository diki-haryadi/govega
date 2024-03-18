package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"

	"bitbucket.org/rctiplus/vegapunk/event"
)

const (
	KafkaPartitionKey     = attribute.Key("messaging.kafka.partition")
	KafkaConsumerGroupKey = attribute.Key("messaging.kafka.consumer_group")
)

type (
	KafkaConsumeMessage struct {
		*KafkaMessageCarrier
		reader *kafka.Reader
	}
)

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
