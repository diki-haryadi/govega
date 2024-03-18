package event

import "context"

type IteratorFunc func(ctx context.Context) (ConsumeMessage, error)

func (fn IteratorFunc) Next(ctx context.Context) (ConsumeMessage, error) {
	return fn(ctx)
}
