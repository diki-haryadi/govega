package mem

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheURL(t *testing.T) {
	urlStr := "mem://"
	u, err := url.Parse(urlStr)
	assert.Nil(t, err)
	assert.NotNil(t, u)

	cache, err := NewCache(u)
	assert.Nil(t, err)
	assert.NotNil(t, cache)

	mc, ok := cache.(*MemoryCache)
	assert.True(t, ok)
	assert.NotNil(t, mc)
}

func TestMemoryCache(t *testing.T) {
	dCache := NewMemoryCache()
	err := dCache.Set(nil, "tesstring", "value", 0)
	assert.Nil(t, err)
	rstring, err := dCache.GetString(nil, "tesstring")
	assert.Nil(t, err)
	assert.Equal(t, "value", rstring)
	rbyte, err := dCache.Get(nil, "tesstring")
	assert.Nil(t, err)
	assert.Equal(t, []byte("value"), rbyte)

	err = dCache.Set(nil, "testint", 123, 0)
	assert.Nil(t, err)
	rint, err := dCache.GetInt(nil, "testint")
	assert.Nil(t, err)
	assert.Equal(t, int64(123), rint)

	err = dCache.Set(nil, "testfloat", 10.5, 0)
	assert.Nil(t, err)
	rfloat, err := dCache.GetFloat(nil, "testfloat")
	assert.Nil(t, err)
	assert.Equal(t, 10.5, rfloat)

	assert.True(t, dCache.Exist(nil, "tesstring"))
	assert.True(t, dCache.Exist(nil, "testint"))

	err = dCache.Set(nil, "testexp", "any", 10)
	assert.Nil(t, err)
	remain := dCache.RemainingTime(nil, "testexp")

	assert.Equal(t, 10, remain)

	time.Sleep(time.Duration(1) * time.Second)
	remain = dCache.RemainingTime(nil, "testexp")
	assert.Equal(t, 9, remain)

}

func TestMemObject(t *testing.T) {

	dCache := NewMemoryCache()
	assert.NotNil(t, dCache)

	obj := map[string]interface{}{
		"env":     "dev",
		"port":    "8080",
		"host":    "localhost",
		"counter": 1,
	}

	err := dCache.Set(nil, "testobj", obj, 0)
	assert.Nil(t, err)

	var res map[string]interface{}

	err = dCache.GetObject(nil, "testobj", &res)
	assert.Nil(t, err)

	assert.Equal(t, obj["env"], res["env"])
	assert.Equal(t, obj["port"], res["port"])
	assert.Equal(t, fmt.Sprintf("%v", obj["counter"]), fmt.Sprintf("%v", res["counter"]))
}

func TestMemBytes(t *testing.T) {

	dCache := NewMemoryCache()
	assert.NotNil(t, dCache)

	err := dCache.Set(nil, "testint", 10, 0)
	assert.Nil(t, err)

	b, err := dCache.Get(nil, "testint")
	assert.Nil(t, err)
	assert.Equal(t, "10", string(b))

	err = dCache.Set(nil, "testbool", true, 0)
	assert.Nil(t, err)

	b, err = dCache.Get(nil, "testbool")
	assert.Nil(t, err)
	assert.Equal(t, "1", string(b))

}
