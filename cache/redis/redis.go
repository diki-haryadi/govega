package redis

import (
	"bitbucket.org/rctiplus/vegapunk/cache"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8"
	"net/url"
	"strings"
	"time"
)

type Cache struct {
	client *redis.Client
	ns     string
}

func NewCache(url *url.URL) (cache.Cache, error) {
	pass, _ := url.User.Password()
	opt := &redis.Options{
		Addr:     url.Host,
		Password: pass,
		DB:       0,
	}

	getTls := url.Query().Get("tls")
	emptyTls := getTls == ""
	if !emptyTls {
		opt.TLSConfig = &tls.Config{ServerName: getTls}
	}

	redisClient := redis.NewClient(opt)
	redisClient.AddHook(apmgoredis.NewHook())

	ns := strings.TrimSuffix(url.Path, "/")
	if ns == "" {
		ns = schemaDefault
	}

	cache := &Cache{
		client: redisClient,
		ns:     strings.TrimPrefix(url.Path, "/"),
	}

	_, err := cache.client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	switch value.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []byte:
		return c.client.
			Set(ctx, c.ns+key, value, time.Duration(expiration)*time.Second).
			Err()
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}

		return c.client.
			Set(ctx, c.ns+key, b, time.Duration(expiration)*time.Second).
			Err()
	}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := c.client.Get(ctx, c.ns+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, cache.ErrNotFound
		}

		return nil, err
	}
	return b, nil
}

func (c *Cache) GetObject(ctx context.Context, key string, doc interface{}) error {
	b, err := c.client.Get(ctx, c.ns+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return cache.ErrNotFound
		}
		return err
	}
	return json.Unmarshal(b, doc)
}

func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	s, err := c.client.Get(ctx, c.ns+key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", cache.ErrNotFound
		}
		return "", err
	}
	return s, nil
}

func (c *Cache) GetInt(ctx context.Context, key string) (int64, error) {
	i, err := c.client.Get(ctx, c.ns+key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, cache.ErrNotFound
		}
		return 0, err
	}
	return i, nil
}

func (c *Cache) GetFloat(ctx context.Context, key string) (float64, error) {
	f, err := c.client.Get(ctx, c.ns+key).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, cache.ErrNotFound
		}
		return 0, err
	}
	return f, nil
}

func (c *Cache) Exist(ctx context.Context, key string) bool {
	return c.client.Exists(ctx, c.ns+key).
		Val() > 0
}

func (c *Cache) Delete(ctx context.Context, key string, opts ...cache.DeleteOptions) error {
	deleteCache := &cache.DeleteCache{}
	for _, opt := range opts {
		opt(deleteCache)
	}

	if deleteCache.Pattern != "" {
		iter := c.client.Scan(ctx, 0, c.ns+deleteCache.Pattern, 0).Iterator()

		var localKeys []string
		for iter.Next(ctx) {
			localKeys = append(localKeys, iter.Val())
		}

		if err := iter.Err(); err != nil {
			return err
		}

		if len(localKeys) > 0 {
			_, err := c.client.Pipelined(ctx, func(p redis.Pipeliner) error {
				p.Del(ctx, localKeys...)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	}

	return c.client.Del(ctx, c.ns+key).Err()
}

func (c *Cache) GetKeys(ctx context.Context, pattern string) []string {
	cmd := c.client.Keys(ctx, pattern)
	keys, err := cmd.Result()
	if err != nil {
		return nil
	}

	return keys
}

func (c *Cache) RemainingTime(ctx context.Context, key string) int {
	return int(c.client.TTL(ctx, c.ns+key).Val().Seconds())
}

func (c *Cache) Close() error {
	return c.client.Close()
}
