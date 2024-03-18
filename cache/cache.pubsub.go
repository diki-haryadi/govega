package cache

import "context"

type Subscriber interface {
	Channel() <-chan string
	ReceiveMessage(ctx context.Context) (string, error)
	Close() error
}
