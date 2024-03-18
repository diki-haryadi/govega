package cache

import (
	"context"

	netUrl "net/url"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration int) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetObject(ctx context.Context, key string, doc interface{}) error
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int64, error)
	GetFloat(ctx context.Context, key string) (float64, error)
	Exist(ctx context.Context, key string) bool
	Delete(ctx context.Context, key string, opts ...DeleteOptions) error
	GetKeys(ctx context.Context, pattern string) []string
	RemainingTime(ctx context.Context, key string) int
	Publish(ctx context.Context, channel, message string) error
	Subscribe(ctx context.Context, topic string) (Subscriber, error)
	Close() error
}

type (
	InitFunc      func(url *netUrl.URL) (Cache, error)
	DeleteOptions func(options *DeleteCache)
)

type DeleteCache struct {
	Pattern string
}

func New(url string) (Cache, error) {
	u, err := netUrl.Parse(url)
	if err != nil {
		return nil, err
	}

	fn, ok := cacheImpl[u.Scheme]
	if !ok {
		return nil, ErrUnsuportedSchema
	}

	return fn(u)
}
