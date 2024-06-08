package pubsub

import (
	"context"
	"os"
	"strings"

	"github.com/diki-haryadi/govega/event"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/kafkapubsub"
)

type PubsubSender struct {
	topics       map[string]*pubsub.Topic
	Schema       string `json:"schema" mapstructure:"schema"`
	KafkaBrokers string `json:"kafka_brokers" mapstructure:"kafka_brokers"`
	propagator   propagation.TextMapPropagator
}

type PubsubMessageCarrier struct {
	*pubsub.Message
}

func init() {
	event.RegisterSender("pubsub", NewPubsubSender)
}

func NewPubsubSender(ctx context.Context, config interface{}) (event.Sender, error) {
	var pub PubsubSender
	if err := mapstructure.Decode(config, &pub); err != nil {
		return nil, err
	}

	if pub.Schema == "kafka://" {
		os.Setenv("KAFKA_BROKERS", pub.KafkaBrokers)
	}
	pub.topics = make(map[string]*pubsub.Topic)
	pub.propagator = otel.GetTextMapPropagator()
	return &pub, nil
}

func (p *PubsubSender) getTopic(ctx context.Context, topicName string) (*pubsub.Topic, error) {
	if top, ok := p.topics[topicName]; ok {
		return top, nil
	}

	var top *pubsub.Topic
	var err error

	switch p.Schema {
	case "kafka://":
		config := kafkapubsub.MinimalConfig()
		top, err = kafkapubsub.OpenTopic(strings.Split(p.KafkaBrokers, ","), config, topicName, &kafkapubsub.TopicOptions{KeyName: "key"})
	default:
		top, err = pubsub.OpenTopic(ctx, p.Schema+topicName)
	}

	if err != nil {
		return nil, err
	}
	p.topics[topicName] = top
	return top, nil
}

func (p *PubsubSender) Send(ctx context.Context, message *event.EventMessage) error {
	tr := otel.Tracer("event/emitter")
	ctx, span := tr.Start(ctx, "SEND "+message.Topic,
		trace.WithAttributes(semconv.MessagingOperationProcess),
		trace.WithAttributes(semconv.MessagingDestinationKindTopic),
		trace.WithAttributes(semconv.MessagingDestinationKey.String(message.Topic)),
		trace.WithAttributes(semconv.MessagingProtocolKey.String(p.Schema)),
		trace.WithAttributes(semconv.MessagingConversationIDKey.String(message.Key)),
	)
	defer span.End()

	topic, err := p.getTopic(ctx, message.Topic)
	if err != nil {
		span.RecordError(err)
		return err
	}
	mb, err := message.ToBytes()

	if err != nil {
		span.RecordError(err)
		return err
	}
	msg := &pubsub.Message{Body: mb}
	if message.Key != "" {
		msg.Metadata = map[string]string{"key": message.Key}
	}
	carrier := newPubsubMessageCarrier(msg)
	p.propagator.Inject(ctx, carrier)

	if err := topic.Send(ctx, msg); err != nil {
		span.RecordError(err)
		return err
	}

	return nil

}

func newPubsubMessageCarrier(msg *pubsub.Message) *PubsubMessageCarrier {
	return &PubsubMessageCarrier{
		Message: msg,
	}
}

// Get retrieves a single value for a given key.
func (k PubsubMessageCarrier) Get(key string) string {
	return k.Metadata[key]
}

// Set sets a header.
func (k PubsubMessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
	k.Metadata[key] = val
}

// Keys returns a slice of all key identifiers in the carrier.
func (k PubsubMessageCarrier) Keys() []string {
	out := make([]string, len(k.Metadata))
	i := 0
	for _, h := range k.Metadata {
		out[i] = h
	}
	return out
}
