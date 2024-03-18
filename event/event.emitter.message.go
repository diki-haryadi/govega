package event

import (
	"context"
	"encoding/json"
)

type (
	Sender interface {
		Send(ctx context.Context, message *EventMessage) error
	}

	Writer interface {
		Sender
		Delete(ctx context.Context, message *EventMessage) error
	}

	EventMessage struct {
		Topic    string                 `json:"-"`
		Key      string                 `json:"-"`
		Data     interface{}            `json:"data,omitempty" mapstructure:"data"`
		Metadata map[string]interface{} `json:"metadata,omitempty" mapstructure:"metadata"`
		RawData  []byte                 `json:"-"` //To provide raw data to consumer
	}
)

type SenderFactory func(ctx context.Context, config interface{}) (Sender, error)
type WriterFactory func(ctx context.Context, config interface{}) (Writer, error)

var (
	senders = map[string]SenderFactory{
		"logger": EventLoggerSender,
	}
	writers = map[string]WriterFactory{
		"logger": EventLoggerWriter,
	}
)

func (m *EventMessage) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *EventMessage) Hash() (string, error) {
	return hash(m)
}
