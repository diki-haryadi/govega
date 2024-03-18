package event

import "context"

func EventLoggerWriter(ctx context.Context, config interface{}) (Writer, error) {
	return NewEventLogger(ctx, config)
}
