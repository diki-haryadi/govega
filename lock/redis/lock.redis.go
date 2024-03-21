package redis

import (
	"context"
	"errors"
	"time"

	"github.com/dikiharyadi19/govegapunk/lock"
	"github.com/dikiharyadi19/govegapunk/lock/redis/redlock"
)

func (l *LockManager) TryLock(ctx context.Context, id string, ttl int) error {
	resource := l.prefix + id
	_, err := l.manager.TryLock(
		ctx,
		resource,
		time.Duration(ttl)*time.Second,
	)

	if errors.Is(err, redlock.ErrAcquireLock) {
		return lock.ErrResourceLocked
	}

	return err
}

func (l *LockManager) Lock(ctx context.Context, id string, ttl int) error {
	resource := l.prefix + id
	_, err := l.manager.Lock(
		ctx,
		resource,
		time.Duration(ttl)*time.Second,
	)

	if errors.Is(err, redlock.ErrAcquireLock) {
		return lock.ErrResourceLocked
	}

	return err
}

func (l *LockManager) Unlock(ctx context.Context, id string) error {
	return l.manager.UnLock(ctx, l.prefix+id)
}

func (l *LockManager) Close() error {
	return l.manager.Close()
}
