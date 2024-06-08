package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cache "github.com/diki-haryadi/govega/cache"
	"github.com/diki-haryadi/govega/cache/embed"
	_ "github.com/diki-haryadi/govega/cache/embed"
	"github.com/diki-haryadi/govega/cache/lru"
	_ "github.com/diki-haryadi/govega/cache/lru"
	"github.com/diki-haryadi/govega/cache/mem"
	_ "github.com/diki-haryadi/govega/cache/mem"
	"github.com/diki-haryadi/govega/cache/redis"
	_ "github.com/diki-haryadi/govega/cache/redis"
	"github.com/stretchr/testify/assert"
)

type sleepFunc func(t time.Duration)

func TestMemCache(t *testing.T) {
	url := "mem://"
	c, err := cache.New(url)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	mc, ok := c.(*mem.MemoryCache)
	assert.True(t, ok)
	assert.NotNil(t, mc)

	testCache(t, c, func(t time.Duration) {
		time.Sleep(t)
	})
}

func TestRedisCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	url := "redis://" + s.Addr()

	c, err := cache.New(url)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	rc, ok := c.(*redis.Cache)
	assert.True(t, ok)
	assert.NotNil(t, rc)

	testCache(t, c, func(t time.Duration) {
		s.FastForward(t)
	})
}

func TestLRUCache(t *testing.T) {
	url := "lru://"
	c, err := cache.New(url)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	mc, ok := c.(*lru.Cache)
	assert.True(t, ok)
	assert.NotNil(t, mc)

	testCache(t, c, func(t time.Duration) {
		time.Sleep(t)
	})
}

func TestEmbedCache(t *testing.T) {
	url := "embed://mem"
	c, err := cache.New(url)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	mc, ok := c.(*embed.BadgerCache)
	assert.True(t, ok)
	assert.NotNil(t, mc)

	testCache(t, c, func(t time.Duration) {
		time.Sleep(t)
	})
}

func testCache(t *testing.T, c cache.Cache, sleep sleepFunc) {
	ctx := context.Background()
	err := c.Set(ctx, "tesstring", "value", 0)
	assert.Nil(t, err)
	rstring, err := c.GetString(ctx, "tesstring")
	assert.Nil(t, err)
	assert.Equal(t, "value", rstring)
	rbyte, err := c.Get(ctx, "tesstring")
	assert.Nil(t, err)
	assert.Equal(t, []byte("value"), rbyte)

	err = c.Set(ctx, "testint", 123, 0)
	assert.Nil(t, err)
	rint, err := c.GetInt(ctx, "testint")
	assert.Nil(t, err)
	assert.Equal(t, int64(123), rint)

	b, err := c.Get(ctx, "testint")
	assert.Nil(t, err)
	assert.Equal(t, "123", string(b))

	err = c.Set(ctx, "testfloat", 10.5, 0)
	assert.Nil(t, err)
	rfloat, err := c.GetFloat(ctx, "testfloat")
	assert.Nil(t, err)
	assert.Equal(t, 10.5, rfloat)

	b, err = c.Get(ctx, "testfloat")
	assert.Nil(t, err)
	assert.Equal(t, "10.5", string(b))

	err = c.Set(ctx, "testbool", true, 0)
	assert.Nil(t, err)

	b, err = c.Get(ctx, "testbool")
	assert.Nil(t, err)
	assert.Equal(t, "1", string(b))

	assert.True(t, c.Exist(ctx, "tesstring"))
	assert.True(t, c.Exist(ctx, "testint"))

	err = c.Set(ctx, "testexp", "any", 10)
	assert.Nil(t, err)
	remain := c.RemainingTime(ctx, "testexp")

	assert.Equal(t, 10, remain)

	sleep(time.Duration(1) * time.Second)

	remain = c.RemainingTime(ctx, "testexp")
	assert.Equal(t, 9, remain)

	obj := map[string]interface{}{
		"env":     "dev",
		"port":    "8080",
		"host":    "localhost",
		"counter": 1,
	}

	err = c.Set(ctx, "testobj", obj, 0)
	assert.Nil(t, err)

	var res map[string]interface{}

	err = c.GetObject(ctx, "testobj", &res)
	assert.Nil(t, err)

	assert.Equal(t, obj["env"], res["env"])
	assert.Equal(t, obj["port"], res["port"])
	assert.Equal(t, fmt.Sprintf("%v", obj["counter"]), fmt.Sprintf("%v", res["counter"]))
}
