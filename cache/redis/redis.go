package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	cache "github.com/diki-haryadi/govega/cache"
	"github.com/go-redis/redis/extra/redisotel"
	redis "github.com/go-redis/redis/v8"
)

const defaultNS = "redis"
const schema = "redis"
const schemaRedisCluster = "redis-cluster"

// Cache redis cache object
type Cache struct {
	client        *redis.Client
	ns            string
	clusterClient *redis.ClusterClient
}

func init() {
	cache.Register(schema, NewCache)
	cache.Register(schemaRedisCluster, NewCacheCluster)
}

// NewCache create new redis cache
func NewCache(url *url.URL) (cache.Cache, error) {
	p, _ := url.User.Password()
	opt := &redis.Options{
		Addr:     url.Host,
		Password: p,
		DB:       0, // use default DB
	}

	if ts := url.Query().Get("tls"); ts != "" {
		opt.TLSConfig = &tls.Config{
			ServerName: ts,
		}
	}

	rClient := redis.NewClient(opt)
	rClient.AddHook(redisotel.TracingHook{})

	ns := strings.TrimPrefix(url.Path, "/")
	if ns == "" {
		ns = defaultNS
	}

	cache := &Cache{
		client: rClient,
		ns:     strings.TrimPrefix(url.Path, "/"),
	}
	_, err := cache.client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func NewCacheCluster(url *url.URL) (cache.Cache, error) {
	address := strings.Split(url.Host, ",")
	if len(address) < 1 {
		return nil, errors.New("invalid address")
	}

	p, _ := url.User.Password()

	opts := &redis.ClusterOptions{
		Addrs:    address,
		Username: url.User.Username(),
		Password: p}

	if ts := url.Query().Get("tls"); ts != "" {
		opts.TLSConfig = &tls.Config{
			ServerName: ts,
		}
	}

	rClient := redis.NewClusterClient(opts)
	rClient.AddHook(redisotel.TracingHook{})

	cache := &Cache{
		clusterClient: rClient,
	}
	_, err := cache.clusterClient.Ping(context.Background()).Result()
	return cache, err
}

// NewRedisCache creating instance of redis cache
func NewRedisCache(ns string, option ...Option) (*Cache, error) {
	r := &redis.Options{}
	for _, o := range option {
		o(r)
	}
	rClient := redis.NewClient(r)
	rClient.AddHook(redisotel.TracingHook{})
	cache := &Cache{
		client: rClient,
		ns:     ns,
	}
	_, err := cache.client.Ping(context.Background()).Result()
	return cache, err
}

func NewRedisCluster(addresses []string, option ...ClusterOption) (*Cache, error) {
	r := &redis.ClusterOptions{}
	for _, o := range option {
		o(r)
	}
	rClient := redis.NewClusterClient(r)
	rClient.AddHook(redisotel.TracingHook{})

	cache := &Cache{
		clusterClient: rClient,
	}
	_, err := cache.clusterClient.Ping(context.Background()).Result()
	return cache, err
}

type Option func(options *redis.Options)

type ClusterOption func(options *redis.ClusterOptions)

func DefaultAddressOption(addresses []string, password string) ClusterOption {
	return func(options *redis.ClusterOptions) {
		options.Password = password
		options.Addrs = addresses
	}
}

func ClusterTLSOption(address string) ClusterOption {
	return func(options *redis.ClusterOptions) {
		options.TLSConfig = &tls.Config{
			ServerName: address,
		}
	}
}

func DefaultOption(address, password string) Option {
	return func(options *redis.Options) {
		options.Password = password
		options.Addr = address
	}
}

func TLSOption(address string) Option {
	return func(options *redis.Options) {
		options.TLSConfig = &tls.Config{
			ServerName: address,
		}
	}
}

// Set set value
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	switch value.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []byte:
		return c.client.Set(ctx, c.ns+key, value, time.Duration(expiration)*time.Second).Err()
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return c.client.Set(ctx, c.ns+key, b, time.Duration(expiration)*time.Second).Err()
	}
}

// Increment increment int value
func (c *Cache) Increment(ctx context.Context, key string, expiration int) (int64, error) {
	switch expiration {
	case 0:
		i, err := c.client.Incr(ctx, key).Result()
		if err != nil {
			return 0, err
		}
		return i, nil
	default:
		pipe := c.client.TxPipeline()

		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Second*time.Duration(expiration))

		_, err := pipe.Exec(ctx)
		if err != nil {
			return 0, err
		}
		return incr.Val(), nil
	}
}

// Get get value
func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := c.client.Get(ctx, c.ns+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, cache.NotFound
		}
		return nil, err
	}
	return b, nil
}

// GetObject get object value
func (c *Cache) GetObject(ctx context.Context, key string, doc interface{}) error {
	b, err := c.client.Get(ctx, c.ns+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return cache.NotFound
		}
		return err
	}
	return json.Unmarshal(b, doc)
}

// GetString get string value
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	s, err := c.client.Get(ctx, c.ns+key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", cache.NotFound
		}
		return "", err
	}
	return s, nil
}

// GetInt get int value
func (c *Cache) GetInt(ctx context.Context, key string) (int64, error) {
	i, err := c.client.Get(ctx, c.ns+key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, cache.NotFound
		}
		return 0, err
	}
	return i, nil
}

// GetFloat get float value
func (c *Cache) GetFloat(ctx context.Context, key string) (float64, error) {
	f, err := c.client.Get(ctx, c.ns+key).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, cache.NotFound
		}
		return 0, err
	}
	return f, nil
}

// Exist check if key exist
func (c *Cache) Exist(ctx context.Context, key string) bool {
	return c.client.Exists(ctx, c.ns+key).Val() > 0
}

// Delete delete record
func (c *Cache) Delete(ctx context.Context, key string, opts ...cache.DeleteOptions) error {
	deleteCache := &cache.DeleteCache{}
	for _, opt := range opts {
		opt(deleteCache)
	}

	if deleteCache.Pattern != "" {
		return c.deletePattern(ctx, deleteCache.Pattern)
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

// deletePattern delete record by pattern
func (c *Cache) deletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, c.ns+pattern, 0).Iterator()
	var localKeys []string

	for iter.Next(ctx) {
		localKeys = append(localKeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(localKeys) > 0 {
		_, err := c.client.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
			pipeline.Del(ctx, localKeys...)
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// RemainingTime get remaining time
func (c *Cache) RemainingTime(ctx context.Context, key string) int {
	return int(c.client.TTL(ctx, c.ns+key).Val().Seconds())
}

// Close close connection
func (c *Cache) Close() error {
	return c.client.Close()
}
