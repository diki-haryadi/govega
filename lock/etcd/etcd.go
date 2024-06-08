package etcd

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/diki-haryadi/govega/lock"
	client "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const schema = "etcd"

type lockSession struct {
	session *concurrency.Session
	lock    *concurrency.Mutex
}

// LockManager ETCD v3 lock manager
type LockManager struct {
	client  *client.Client
	prefix  string
	session map[string]lockSession
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

	cli, err := client.New(client.Config{
		Endpoints:   hs,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &LockManager{
		client:  cli,
		prefix:  urls[0].Path,
		session: make(map[string]lockSession),
	}, nil
}

// TryLock try to lock, and return immediately if resource already locked
func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	s, err := concurrency.NewSession(l.client, concurrency.WithTTL(5))
	if err != nil {
		return err
	}
	//defer s.Close()
	el := concurrency.NewMutex(s, l.prefix+id)

	if err := el.TryLock(ctx); err != nil {
		s.Close()
		if errors.Is(err, concurrency.ErrLocked) {
			return lock.ErrResourceLocked
		}

		return err
	}

	l.session[id] = lockSession{session: s, lock: el}
	return nil
}

// Lock try to lock and wait until resource is available to lock
func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	s, err := concurrency.NewSession(l.client, concurrency.WithTTL(5))
	if err != nil {
		return err
	}
	//defer s.Close()
	el := concurrency.NewMutex(s, l.prefix+id)

	if err := el.Lock(ctx); err != nil {
		s.Close()
		return err
	}

	l.session[id] = lockSession{session: s, lock: el}
	return nil
}

// Unlock unlock resource
func (l *LockManager) Unlock(ctx context.Context, id string) error {
	s := l.session[id]

	if s.lock != nil {
		if err := s.lock.Unlock(ctx); err != nil {
			return err
		}
	}
	if s.session != nil {
		if err := s.session.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Close close client
func (l *LockManager) Close() error {
	return l.client.Close()
}
