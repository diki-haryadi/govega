package event

import "context"

func EventLoggerListener(ctx context.Context, config interface{}) (Listener, error) {
	return NewEventLogger(ctx, config)
}
