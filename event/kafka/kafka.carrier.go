package kafka

import "github.com/segmentio/kafka-go"

type (
	KafkaMessageCarrier struct {
		*kafka.Message
	}
)

func newKafkaMessageCarrier(msg *kafka.Message) *KafkaMessageCarrier {
	return &KafkaMessageCarrier{
		Message: msg,
	}
}

// Get retrieves a single value for a given key.
func (k KafkaMessageCarrier) Get(key string) string {
	for _, h := range k.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set sets a header.
func (k KafkaMessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
	for i := 0; i < len(k.Headers); i++ {
		if k.Headers[i].Key == key {
			k.Headers = append(k.Headers[:i], k.Headers[i+1:]...)
			i--
		}
	}
	k.Headers = append(k.Headers, kafka.Header{
		Key:   key,
		Value: []byte(val),
	})
}

// Keys returns a slice of all key identifiers in the carrier.
func (k KafkaMessageCarrier) Keys() []string {
	out := make([]string, len(k.Headers))
	for i, h := range k.Headers {
		out[i] = string(h.Key)
	}
	return out
}
