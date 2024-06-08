package redis

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/diki-haryadi/govega/cache"
	"github.com/stretchr/testify/assert"
)

func TestCacheURL(t *testing.T) {

	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	urlStr := "redis://" + s.Addr()
	u, err := url.Parse(urlStr)
	assert.Nil(t, err)
	assert.NotNil(t, u)

	cache, err := NewCache(u)
	assert.Nil(t, err)
	assert.NotNil(t, cache)

	rc, ok := cache.(*Cache)
	assert.True(t, ok)
	assert.NotNil(t, rc)
}

func TestRedisCache(t *testing.T) {
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	dCache, err := NewRedisCache("", DefaultOption(s.Addr(), ""))
	assert.NotNil(t, dCache)
	assert.Nil(t, err)
	err = dCache.Set(ctx, "tesstring", "value", 0)
	assert.Nil(t, err)
	rstring, err := dCache.GetString(ctx, "tesstring")
	assert.Nil(t, err)
	assert.Equal(t, "value", rstring)

	err = dCache.Set(ctx, "testint", 123, 0)
	assert.Nil(t, err)
	rint, err := dCache.GetInt(ctx, "testint")
	assert.Nil(t, err)
	assert.Equal(t, int64(123), rint)

	rint, err = dCache.Increment(ctx, "testincr", 0)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), rint)

	rint, err = dCache.Increment(ctx, "testincr", 10)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), rint)

	remain := dCache.RemainingTime(ctx, "testincr")
	assert.Equal(t, 10, remain)

	err = dCache.Set(ctx, "testfloat", 10.5, 0)
	assert.Nil(t, err)
	rfloat, err := dCache.GetFloat(ctx, "testfloat")
	assert.Nil(t, err)
	assert.Equal(t, 10.5, rfloat)

	assert.True(t, dCache.Exist(ctx, "tesstring"))
	assert.True(t, dCache.Exist(ctx, "testint"))

	err = dCache.Set(ctx, "testexp", "any", 10)
	assert.Nil(t, err)
	remain = dCache.RemainingTime(ctx, "testexp")

	assert.Equal(t, 10, remain)

	s.FastForward(time.Duration(1) * time.Second)

	//time.Sleep(time.Duration(1) * time.Second)
	remain = dCache.RemainingTime(ctx, "testexp")
	assert.Equal(t, 9, remain)

	err = dCache.Delete(ctx, "tesstring")
	assert.Nil(t, err)
	rstring, err = dCache.GetString(ctx, "tesstring")
	assert.Equal(t, cache.NotFound, err)
	assert.Equal(t, "", rstring)
}

func TestRedisObject(t *testing.T) {
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	dCache, err := NewRedisCache("", DefaultOption(s.Addr(), ""))
	assert.NotNil(t, dCache)
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"env":     "dev",
		"port":    "8080",
		"host":    "localhost",
		"counter": 1,
	}

	err = dCache.Set(ctx, "testobj", obj, 0)
	assert.Nil(t, err)

	var res map[string]interface{}

	err = dCache.GetObject(ctx, "testobj", &res)
	assert.Nil(t, err)

	assert.Equal(t, obj["env"], res["env"])
	assert.Equal(t, obj["port"], res["port"])
	assert.Equal(t, fmt.Sprintf("%v", obj["counter"]), fmt.Sprintf("%v", res["counter"]))
}

func TestRedisBytes(t *testing.T) {
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	dCache, err := NewRedisCache("", DefaultOption(s.Addr(), ""))
	assert.NotNil(t, dCache)
	assert.Nil(t, err)

	err = dCache.Set(ctx, "testint", 10, 0)
	assert.Nil(t, err)

	b, err := dCache.Get(ctx, "testint")
	assert.Nil(t, err)
	assert.Equal(t, "10", string(b))

	err = dCache.Set(ctx, "testbool", true, 0)
	assert.Nil(t, err)

	b, err = dCache.Get(ctx, "testbool")
	assert.Nil(t, err)
	assert.Equal(t, "1", string(b))

}

func TestRedisGetKeys(t *testing.T) {
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	dCache, err := NewRedisCache("", DefaultOption(s.Addr(), ""))
	assert.NotNil(t, dCache)
	assert.Nil(t, err)

	err = dCache.Set(ctx, "testing", 10, 0)
	assert.Nil(t, err)

	err = dCache.Set(ctx, "testin", 10, 0)
	assert.Nil(t, err)

	err = dCache.Set(ctx, "testi", 10, 0)
	assert.Nil(t, err)

	err = dCache.Set(ctx, "test", 10, 0)
	assert.Nil(t, err)

	b := dCache.GetKeys(ctx, "*test*")
	assert.Nil(t, err)
	assert.Equal(t, []string{"test", "testi", "testin", "testing"}, b)

}
