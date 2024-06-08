package lock

import (
	"context"
	"sync"
	"time"
)

const interval = 100

// LockManager local lock manager
type LockManager struct {
	mux    *sync.Mutex
	locked map[string]time.Time
}

// New create redis locker instance
func Local() (DLocker, error) {
	return &LockManager{
		mux:    &sync.Mutex{},
		locked: make(map[string]time.Time),
	}, nil
}

func (l *LockManager) lock(id string, ttl int) error {
	l.mux.Lock()
	defer l.mux.Unlock()
	t, ok := l.locked[id]
	if ok && t.After(time.Now()) {
		return ErrResourceLocked
	}

	l.locked[id] = time.Now().Add(time.Duration(ttl) * time.Second)

	return nil
}

// TryLock try to lock, and return immediately if resource already locked
func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	return l.lock(id, ttl)
}

// Lock try to lock and wait until resource is available to lock
func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	err := l.lock(id, ttl)
	if err == nil {
		return nil
	}

	count := 0
	max := ttl * 1000 / interval
	for {
		time.Sleep(time.Duration(interval) * time.Millisecond)
		err := l.lock(id, ttl)
		if err == nil {
			return nil
		}
		count++
		if count > max {
			return err
		}
	}
}

// Unlock unlock resource
func (l *LockManager) Unlock(ctx context.Context, id string) error {
	l.mux.Lock()
	defer l.mux.Unlock()
	delete(l.locked, id)
	return nil
}

// Close close the lock
func (l LockManager) Close() error {
	return nil
}
