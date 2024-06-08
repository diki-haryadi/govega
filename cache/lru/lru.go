package lru

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	cache "github.com/diki-haryadi/govega/cache"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mitchellh/mapstructure"
)

const schema = "lru"
const defaultSize = 1024

type object struct {
	expired time.Time
	value   interface{}
}

// Cache lru cache object
type Cache struct {
	size   int
	scaled bool
	data   *lru.Cache
}

func init() {
	cache.Register(schema, NewCache)
}

// NewCache create new memory cache
func NewCache(url *url.URL) (cache.Cache, error) {
	path := strings.TrimPrefix(url.Path, "/")
	s, err := strconv.Atoi(path)
	if err != nil {
		s = defaultSize
	}

	c, err := lru.New(s)
	if err != nil {
		return nil, err
	}
	return &Cache{
		data:   c,
		size:   s,
		scaled: false,
	}, nil
}

// NewLRUCache new lru instance
func NewLRUCache() *Cache {
	c, err := lru.New(defaultSize)
	if err != nil {
		return nil
	}
	return &Cache{
		data:   c,
		size:   defaultSize,
		scaled: false,
	}
}

func (c *Cache) set(key string, value interface{}, exp int) {
	mo := object{value: value}
	if exp > 0 {
		mo.expired = time.Now().Add(time.Duration(exp) * time.Second)
	}

	if e := c.data.Add(key, mo); e {
		if !c.scaled {
			c.data.Resize(c.size * 2)
		}
	}
}

func (c *Cache) get(key string) interface{} {

	ob, ok := c.data.Get(key)
	if !ok {
		return nil
	}

	val, ok := ob.(object)
	if !ok {
		return ob
	}

	if !val.expired.IsZero() && time.Now().After(val.expired) {
		c.data.Remove(key)
		return nil
	}

	return val.value
}

// Set set value
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	c.set(key, value, expiration)
	return nil
}

func (c *Cache) Increment(ctx context.Context, key string, expiration int) (int64, error) {
	return 0, cache.NotSupported
}

// Get get value
func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	val := c.get(key)
	if val == nil {
		return nil, cache.NotFound
	}

	switch val := val.(type) {
	case int, int8, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return []byte(fmt.Sprintf("%v", val)), nil
	case bool:
		if val {
			return []byte("1"), nil
		}
		return []byte("0"), nil
	case string:
		return []byte(val), nil
	default:
		return json.Marshal(val)
	}
}

// GetObject get value in object
func (c *Cache) GetObject(ctx context.Context, key string, doc interface{}) error {
	val := c.get(key)
	if val == nil {
		return cache.NotFound
	}

	return mapstructure.Decode(val, doc)
}

// GetString get string value
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	val := c.get(key)
	if val == nil {
		return "", cache.NotFound
	}
	sval, sok := val.(string)
	if sok {
		return sval, nil
	}
	return "", errors.New("invalid stored value")
}

// GetInt get int value
func (c *Cache) GetInt(ctx context.Context, key string) (int64, error) {
	val := c.get(key)
	if val == nil {
		return 0, cache.NotFound
	}

	vi, err := strconv.Atoi(fmt.Sprintf("%v", val))
	if err != nil {
		return 0, err
	}

	return int64(vi), nil
}

// GetFloat get float value
func (c *Cache) GetFloat(ctx context.Context, key string) (float64, error) {
	val := c.get(key)
	if val == nil {
		return 0, cache.NotFound
	}

	f, ok := val.(float64)
	if !ok {
		return 0, errors.New("invalid stored value")
	}

	return f, nil
}

// Exist check if key exist
func (c *Cache) Exist(ctx context.Context, key string) bool {
	return c.data.Contains(key)
}

func (c *Cache) GetKeys(ctx context.Context, pattern string) []string {
	// TODO implement me
	return nil
}

// RemainingTime get remainig time
func (c *Cache) RemainingTime(ctx context.Context, key string) int {

	ob, ok := c.data.Get(key)
	if !ok {
		return -1
	}

	val, ok := ob.(object)
	if !ok {
		return -1
	}

	if !val.expired.IsZero() && time.Now().After(val.expired) {
		c.data.Remove(key)
		return 0
	}

	if val.expired.IsZero() {
		return 0
	}

	//return int(time.Until(val.expired))
	return int(math.Ceil(time.Until(val.expired).Seconds()))
}

// Delete delete record
func (c *Cache) Delete(ctx context.Context, key string, opts ...cache.DeleteOptions) error {
	c.data.Remove(key)
	return nil
}

// Close close cache
func (c *Cache) Close() error {
	c.data, _ = lru.New(c.size)
	return nil
}
