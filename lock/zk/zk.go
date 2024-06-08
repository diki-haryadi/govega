package zk

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/diki-haryadi/govega/lock"
	"github.com/diki-haryadi/govega/lock/pkg/zklock"
)

const schema = "zk"

// LockManager ZooKeeper Lock manager
type LockManager struct {
	prefix  string
	session map[string]*zklock.Dlocker
}

func init() {
	lock.Register(schema, New)
}

// New new etcd locker
func New(urls []*url.URL) (lock.DLocker, error) {
	hs := make([]string, 0)
	for _, u := range urls {
		hs = append(hs, u.Host)
	}

	if err := zklock.Connect(hs, 20*time.Second); err != nil {
		return nil, err
	}

	return &LockManager{
		prefix:  urls[0].Path,
		session: make(map[string]*zklock.Dlocker),
	}, nil
}

// TryLock try to lock, and return immediately if resource already locked
func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	zl, err := zklock.NewLocker(l.prefix+id, time.Duration(ttl)*time.Second)
	if err != nil {
		return err
	}

	if err := zl.TryLock(); err != nil {
		if errors.Is(err, zklock.ErrAcquireLock) {
			return lock.ErrResourceLocked
		}

		return err
	}
	l.session[id] = zl
	return nil
}

// Lock try to lock and wait until resource is available to lock
func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	zl, err := zklock.NewLocker(l.prefix+id, time.Duration(ttl)*time.Second)
	if err != nil {
		return err
	}

	if err := zl.Lock(); err != nil {
		return err
	}
	l.session[id] = zl
	return nil
}

// Unlock unlock resource
func (l *LockManager) Unlock(ctx context.Context, id string) error {
	s := l.session[id]

	if s != nil {
		return s.Unlock()
	}
	return nil
}

// Close close client
func (l *LockManager) Close() error {
	zklock.Close()
	return nil
}
