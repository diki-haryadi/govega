package event

import "context"

func EventLoggerSender(ctx context.Context, config interface{}) (Sender, error) {
	return NewEventLogger(ctx, config)
}
