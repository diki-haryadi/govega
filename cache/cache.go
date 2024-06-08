package cache

import (
	"context"
	"errors"
	"net/url"
)

const NotFound = CacheError("[cache] not found")
const NotSupported = CacheError("[cache] not supported")

// Cache cache service
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration int) error
	Increment(ctx context.Context, key string, expiration int) (int64, error)
	Get(ctx context.Context, key string) ([]byte, error)
	GetObject(ctx context.Context, key string, doc interface{}) error
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int64, error)
	GetFloat(ctx context.Context, key string) (float64, error)
	Exist(ctx context.Context, key string) bool
	Delete(ctx context.Context, key string, opts ...DeleteOptions) error
	GetKeys(ctx context.Context, pattern string) []string
	RemainingTime(ctx context.Context, key string) int
	Close() error
}

type DeleteCache struct {
	Pattern string
}

type DeleteOptions func(options *DeleteCache)

// InitFunc cache init function
type InitFunc func(url *url.URL) (Cache, error)

var cacheImpl = make(map[string]InitFunc)

// Register register cache implementation
func Register(schema string, f InitFunc) {
	cacheImpl[schema] = f
}

// New create new cache
func New(urlStr string) (Cache, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	f, ok := cacheImpl[u.Scheme]
	if !ok {
		return nil, errors.New("unsupported scheme")
	}

	return f(u)
}

type CacheError string

func (e CacheError) Error() string { return string(e) }
