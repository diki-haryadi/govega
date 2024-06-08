package lock

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrResourceLocked = errors.New("resource locked")
)

// DLocker distributed locker interface
type DLocker interface {
	TryLock(ctx context.Context, id string, ttl int) error
	Lock(ctx context.Context, id string, ttl int) error
	Unlock(ctx context.Context, id string) error
	Close() error
}

// InitFunc cache init function
type InitFunc func(urls []*url.URL) (DLocker, error)

var lockerImpl = make(map[string]InitFunc)

// Register register cache implementation
func Register(schema string, f InitFunc) {
	lockerImpl[schema] = f
}

// New create new cache
func New(urlStr string) (DLocker, error) {
	if urlStr == "" {
		urlStr = "local://"
	}

	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	urls := strings.Split(urlStr, ",")
	if len(urls) > 1 {
		first, err := url.Parse(urls[0])
		if err != nil {
			return nil, err
		}
		scheme := first.Scheme + "://"

		last, err := url.Parse(scheme + urls[len(urls)-1])
		if err != nil {
			return nil, err
		}
		path := last.Path

		for i, u := range urls {
			if !strings.HasPrefix(u, scheme) {
				u = scheme + u
			}
			if !strings.HasSuffix(u, path) {
				u += path
			}
			urls[i] = u
		}
	}

	up := make([]*url.URL, 0)
	for _, us := range urls {
		u, err := url.Parse(us)
		if err != nil {
			return nil, err
		}
		up = append(up, u)
	}

	if up[0].Scheme == "local" {
		return Local()
	}

	f, ok := lockerImpl[up[0].Scheme]
	if !ok {
		return nil, errors.New("unsupported scheme")
	}

	return f(up)
}
