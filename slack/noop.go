package slack

import (
	"context"

	"github.com/diki-haryadi/govega/log"
)

func (n *noopSlack) PostToSlack(ctx context.Context, message string, opts ...PostOption) error {
	payload := payload{
		Fields: map[string]interface{}{},
	}

	for _, o := range opts {
		o(&payload)
	}

	fields := log.Fields{
		"message": message,
	}

	if payload.StackTrace != "" {
		fields["stacktrac"] = payload.StackTrace
	}

	for k, v := range payload.Fields {
		fields[k] = v
	}

	log.WithFields(fields).WithContext(ctx).Println("[slack/noopslack] new post")
	return nil
}
