package lock

import (
	"context"
	"sync"
	"time"
)

const (
	interval = 100
)

type LockManager struct {
	mux    *sync.Mutex
	locked map[string]time.Time
}

func Local() (DLocker, error) {
	return &LockManager{
		mux:    &sync.Mutex{},
		locked: make(map[string]time.Time),
	}, nil
}

func (l *LockManager) lock(id string, ttl int) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	if t, ok := l.locked[id]; ok && t.After(time.Now()) {
		return ErrResourceLocked
	}

	l.locked[id] = time.Now().
		Add(time.Duration(ttl) * time.Second)
	return nil
}

func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	return l.lock(id, ttl)
}

func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	err := l.lock(id, ttl)
	if err == nil {
		return nil
	}

	count := 0
	max := ttl * 1000 / interval
	for {
		time.Sleep(time.Duration(interval) * time.Microsecond)

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

func (l *LockManager) Unlock(ctx context.Context, id string) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	delete(l.locked, id)
	return nil
}

func (l *LockManager) Close() error {
	return nil
}
