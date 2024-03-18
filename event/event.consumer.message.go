package event

import (
	"context"
	"encoding/json"
	"fmt"
)

type (
	EventConsumeMessage struct {
		Topic    string
		Key      string
		Metadata map[string]interface{}
		Data     []byte
	}

	eventConsumeMessageRead struct {
		Data     json.RawMessage        `json:"data" mapstructure:"data"`
		Metadata map[string]interface{} `json:"metadata,omitempty" mapstructure:"metadata"`
	}
)

//NewEventConsumeMessage return event consume message from byte data
func NewEventConsumeMessage(v []byte) (*EventConsumeMessage, error) {
	if v == nil {
		return &EventConsumeMessage{}, nil
	}

	var readmsg eventConsumeMessageRead
	if err := json.Unmarshal(v, &readmsg); err != nil {
		return nil, fmt.Errorf("[event/consumer] failed to unmarshal value: %w", err)
	}

	return &EventConsumeMessage{
		Metadata: readmsg.Metadata,
		Data:     readmsg.Data,
	}, nil
}

//GetConsumerGroupFromContext return consumer group from contet if any
func GetConsumerGroupFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(consumerGroupKey).(string); ok {
		return value
	}
	return ""
}
