package mem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	cache "github.com/diki-haryadi/govega/cache"
	"github.com/mitchellh/mapstructure"
)

const schema = "mem"

type memObject struct {
	expired time.Time
	value   interface{}
}

// MemoryCache memory cache object
type MemoryCache struct {
	data map[string]memObject
	mux  *sync.RWMutex
	//lifetime map[string]time.Time
}

func init() {
	cache.Register(schema, NewCache)
}

// NewCache create new memory cache
func NewCache(url *url.URL) (cache.Cache, error) {
	return &MemoryCache{
		data: make(map[string]memObject),
		mux:  &sync.RWMutex{},
	}, nil
}

// NewMemoryCache new memory instance
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]memObject),
		mux:  &sync.RWMutex{},
	}
}

func (m *MemoryCache) set(key string, value interface{}, exp int) {
	m.mux.Lock()
	mo := memObject{value: value}
	if exp > 0 {
		mo.expired = time.Now().Add(time.Duration(exp) * time.Second)
	}
	m.data[key] = mo
	m.mux.Unlock()
}

func (m *MemoryCache) get(key string) interface{} {
	m.mux.RLock()
	val, ok := m.data[key]
	m.mux.RUnlock()
	if !ok {
		return nil
	}

	if !val.expired.IsZero() && time.Now().After(val.expired) {
		m.del(key)
		return nil
	}

	return val.value
}

func (m *MemoryCache) del(key string) {
	m.mux.Lock()
	delete(m.data, key)
	m.mux.Unlock()
}

// Set set value
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	m.set(key, value, expiration)
	return nil
}

func (m *MemoryCache) Increment(ctx context.Context, key string, expiration int) (int64, error) {
	return 0, cache.NotSupported
}

// Get get value
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	val := m.get(key)
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
func (m *MemoryCache) GetObject(ctx context.Context, key string, doc interface{}) error {
	val := m.get(key)
	if val == nil {
		return cache.NotFound
	}

	return mapstructure.Decode(val, doc)
}

// GetString get string value
func (m *MemoryCache) GetString(ctx context.Context, key string) (string, error) {
	val := m.get(key)
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
func (m *MemoryCache) GetInt(ctx context.Context, key string) (int64, error) {
	val := m.get(key)
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
func (m *MemoryCache) GetFloat(ctx context.Context, key string) (float64, error) {
	val := m.get(key)
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
func (m *MemoryCache) Exist(ctx context.Context, key string) bool {
	return m.get(key) != nil
}

// RemainingTime get remainig time
func (m *MemoryCache) RemainingTime(ctx context.Context, key string) int {
	m.mux.RLock()
	val, ok := m.data[key]
	m.mux.RUnlock()
	if !ok {
		return -1
	}

	if !val.expired.IsZero() && time.Now().After(val.expired) {
		m.del(key)
		return 0
	}

	if val.expired.IsZero() {
		return 0
	}

	return int(math.Ceil(time.Until(val.expired).Seconds()))
}

// Delete delete record
func (m *MemoryCache) Delete(ctx context.Context, key string, opts ...cache.DeleteOptions) error {
	m.del(key)
	return nil
}

func (m *MemoryCache) GetKeys(ctx context.Context, pattern string) []string {
	// TODO implement me
	return nil
}

// Close close cache
func (m *MemoryCache) Close() error {
	m.mux.Lock()
	m.data = make(map[string]memObject)
	m.mux.Unlock()
	return nil
}
