package redis

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/diki-haryadi/govega/lock"
	"github.com/diki-haryadi/govega/lock/pkg/redlock"
)

const schema = "redis"

// LockManager Redis lock manager
type LockManager struct {
	manager *redlock.RedLock
	prefix  string
}

func init() {
	lock.Register(schema, New)
}

// New create redis locker instance
func New(urls []*url.URL) (lock.DLocker, error) {
	hs := make([]string, 0)
	for _, u := range urls {
		host := "tcp://"

		pass, ok := u.User.Password()
		if ok {
			host += u.User.Username() + ":" + pass + "@"
		}

		host += u.Host
		hs = append(hs, host)
	}

	lockMgr, err := redlock.NewRedLock(hs)
	if err != nil {
		return nil, err
	}

	path := strings.ReplaceAll(urls[0].Path, "/", "")
	if !strings.HasSuffix(path, ":") {
		path += ":"
	}

	return &LockManager{
		manager: lockMgr,
		prefix:  path,
	}, nil
}

// TryLock try to lock, and return immediately if resource already locked
func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	_, err := l.manager.TryLock(ctx, l.prefix+id, time.Duration(ttl)*time.Second)
	if errors.Is(err, redlock.ErrAcquireLock) {
		return lock.ErrResourceLocked
	}
	return err
}

// Lock try to lock and wait until resource is available to lock
func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	_, err := l.manager.Lock(ctx, l.prefix+id, time.Duration(ttl)*time.Second)
	if errors.Is(err, redlock.ErrAcquireLock) {
		return lock.ErrResourceLocked
	}
	return err
}

// Unlock unlock resource
func (l *LockManager) Unlock(ctx context.Context, id string) error {
	return l.manager.UnLock(ctx, l.prefix+id)
}

// Close close the lock
func (l *LockManager) Close() error {
	return l.manager.Close()
}
