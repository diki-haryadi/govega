package redlock

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

var redisServers = []string{
	"tcp://127.0.0.1:6379",
	"tcp://127.0.0.1:6380",
	"tcp://127.0.0.1:6381",
}

func TestBasicLock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	lock, err := NewRedLock(redisServers)
	assert.Nil(t, err)

	_, err = lock.Lock(ctx, "foo", 200*time.Millisecond)
	assert.Nil(t, err)
	err = lock.UnLock(ctx, "foo")
	assert.Nil(t, err)
}

func TestUnlockExpiredKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	lock, err := NewRedLock(redisServers)
	assert.Nil(t, err)

	_, err = lock.Lock(ctx, "foo", 50*time.Millisecond)
	assert.Nil(t, err)
	time.Sleep(51 * time.Millisecond)
	err = lock.UnLock(ctx, "foo")
	assert.Nil(t, err)
}

const (
	fpath = "./counter.log"
)

func writer(count int, back chan *countResp) {
	ctx := context.Background()
	lock, err := NewRedLock(redisServers)

	if err != nil {
		back <- &countResp{
			err: err,
		}
		return
	}

	incr := 0
	for i := 0; i < count; i++ {
		expiry, err := lock.Lock(ctx, "foo", 1000*time.Millisecond)
		if err != nil {
			log.Println(err)
		} else {
			if expiry > 500 {
				f, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
				if err != nil {
					back <- &countResp{
						err: err,
					}
					return
				}

				buf := make([]byte, 1024)
				n, _ := f.Read(buf)
				num, _ := strconv.ParseInt(strings.TrimRight(string(buf[:n]), "\n"), 10, 64)
				f.WriteAt([]byte(strconv.Itoa(int(num+1))), 0)
				incr++

				f.Sync()
				f.Close()

				lock.UnLock(ctx, "foo")
			}
		}
	}
	back <- &countResp{
		count: incr,
		err:   nil,
	}
}

func init() {
	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	f.WriteString("0")
	defer f.Close()
}

type countResp struct {
	count int
	err   error
}

func TestSimpleCounter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	routines := 5
	inc := 100
	total := 0
	done := make(chan *countResp, routines)
	for i := 0; i < routines; i++ {
		go writer(inc, done)
	}
	for i := 0; i < routines; i++ {
		resp := <-done
		assert.Nil(t, resp.err)
		total += resp.count
	}

	f, err := os.OpenFile(fpath, os.O_RDONLY, os.ModePerm)
	assert.Nil(t, err)
	defer f.Close()
	buf := make([]byte, 1024)
	n, _ := f.Read(buf)
	counterInFile, _ := strconv.Atoi(string(buf[:n]))
	assert.Equal(t, total, counterInFile)
}

func TestParseConnString(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testCases := []struct {
		addr    string
		success bool
		opts    *redis.Options
	}{
		{"127.0.0.1", false, nil},
		{"127.0.0.1:6379", false, nil}, // must provide scheme
		{"tcp://127.0.0.1:6379", true, &redis.Options{Addr: "127.0.0.1:6379"}},
		{"tcp://:password@127.0.0.1:6379/2?DialTimeout=1.5&ReadTimeout=2&WriteTimeout=2", false, nil},
		{"tcp://:password@127.0.0.1:6379/2?DialTimeout=1&ReadTimeout=2.5&WriteTimeout=2", false, nil},
		{"tcp://:password@127.0.0.1:6379/2?DialTimeout=1&ReadTimeout=2&WriteTimeout=2.5", false, nil},
		{"tcp://:password@127.0.0.1:6379/2?DialTimeout=1&ReadTimeout=2&WriteTimeout=2",
			true, &redis.Options{
				Addr: "127.0.0.1:6379", Password: "password", DB: 2,
				DialTimeout: 1, ReadTimeout: 2, WriteTimeout: 2}},
	}
	for _, tc := range testCases {
		opts, err := parseConnString(tc.addr)
		if tc.success {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
			assert.Exactly(t, tc.opts, opts)
		}
	}
}

func TestNewRedLockError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testCases := []struct {
		addrs   []string
		success bool
	}{
		{[]string{"127.0.0.1:6379"}, false},
		{[]string{"tcp://127.0.0.1:6379", "tcp://127.0.0.1:6380"}, false},
		{[]string{"tcp://127.0.0.1:6379", "tcp://127.0.0.1:6380", "tcp://127.0.0.1:6381"}, true},
	}
	for _, tc := range testCases {
		_, err := NewRedLock(tc.addrs)
		if tc.success {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}

func TestRedlockSetter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	lock, err := NewRedLock(redisServers)
	assert.Nil(t, err)

	retryCount := lock.retryCount
	lock.SetRetryCount(0)
	assert.Equal(t, retryCount, lock.retryCount)
	lock.SetRetryCount(retryCount + 3)
	assert.Equal(t, retryCount+3, lock.retryCount)

	retryDelay := lock.retryDelay
	lock.SetRetryDelay(0)
	assert.Equal(t, retryDelay, lock.retryDelay)
	lock.SetRetryDelay(retryDelay + 100)
	assert.Equal(t, retryDelay+100, lock.retryDelay)
}

func TestAcquireLockFailed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	servers := make([]string, 0, len(redisServers))
	clis := make([]*redis.Client, 0, len(redisServers))
	for _, server := range redisServers {
		server2 := fmt.Sprintf("%s/3?DialTimeout=1&ReadTimeout=1&WriteTimeout=1", server)
		servers = append(servers, server2)
		opts, err := parseConnString(server2)
		assert.Nil(t, err)
		clis = append(clis, redis.NewClient(opts))
	}
	var wg sync.WaitGroup
	for idx, cli := range clis {
		// block two of redis instances
		if idx == 0 {
			continue
		}
		wg.Add(1)
		go func(c *redis.Client) {
			defer wg.Done()
			dur := 4 * time.Second
			c.ClientPause(ctx, dur)
			time.Sleep(dur)
		}(cli)
	}
	lock, err := NewRedLock(servers)
	assert.Nil(t, err)

	validity, err := lock.Lock(ctx, "foo", 100*time.Millisecond)
	assert.Equal(t, int64(0), validity)
	assert.NotNil(t, err)

	wg.Wait()
}

func TestLockContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithCancel(context.Background())

	clis := make([]*redis.Client, 0, len(redisServers))
	for _, server := range redisServers {
		opts, err := parseConnString(server)
		assert.Nil(t, err)
		clis = append(clis, redis.NewClient(opts))
	}
	var wg sync.WaitGroup
	for _, cli := range clis {
		wg.Add(1)
		go func(c *redis.Client) {
			defer wg.Done()
			c.ClientPause(ctx, time.Second)
			time.Sleep(time.Second)
		}(cli)
	}
	lock, err := NewRedLock(redisServers)
	assert.Nil(t, err)

	cancel()
	_, err = lock.Lock(ctx, "foo", 100*time.Millisecond)
	assert.Equal(t, err, context.Canceled)
	wg.Wait()
}

func testKVCacheWrap(t *testing.T, cacheType string) {
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock, err := NewRedLock(redisServers)
			assert.Nil(t, err)
			lock.SetCache(cacheType, nil)
			for j := 0; j < 100; j++ {
				_, err = lock.Lock(ctx, "foo", 200*time.Millisecond)
				assert.Nil(t, err)
				err = lock.UnLock(ctx, "foo")
				assert.Nil(t, err)
			}
			assert.Zero(t, lock.cache.Size())
		}()
	}
	wg.Wait()
}

func TestKVCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testKVCacheWrap(t, CacheTypeSimple)
	testKVCacheWrap(t, CacheTypeFreeCache)
}
