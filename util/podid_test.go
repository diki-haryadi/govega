package util

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/diki-haryadi/govega/cache"
	_ "github.com/diki-haryadi/govega/cache/mem"
	"github.com/diki-haryadi/govega/lock"
)

func TestDistributedPodID_concurrentNoDuplicate(t *testing.T) {
	c, err := cache.New("mem://")
	if err != nil {
		t.Fatalf("failed to init cache: %v", err)
	}

	l, err := lock.Local()
	if err != nil {
		t.Fatalf("failed to init lock: %v", err)
	}

	total := 100

	type pidResult struct {
		id  int64
		err error
	}
	resultChan := make(chan pidResult, 10)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < total; i++ {
		num := i
		go func() {
			pid, err := initTestDistrbutedPodID(ctx, "test", strconv.Itoa(num), c, l)
			if err != nil {
				resultChan <- pidResult{err: fmt.Errorf("failed to init pod id [%d]: %w", num, err)}
				return
			}
			resultChan <- pidResult{id: pid.ID()}
		}()
	}

	m := map[int64]bool{}
	dup := map[int64]bool{}
	for i := 0; i < total; i++ {
		result := <-resultChan
		if result.err != nil {
			t.Fatal(result.err)
		}

		if m[result.id] {
			dup[result.id] = true
			continue
		}

		m[result.id] = true
	}

	if len(dup) > 0 {
		arr := make([]int64, 0)
		for k := range dup {
			arr = append(arr, k)
		}

		t.Fatalf("found duplicate id [%v]", arr)
	}
}

func TestDistributedPodID_reuseIDAfterRelease(t *testing.T) {
	c, err := cache.New("mem://")
	if err != nil {
		t.Fatalf("failed to init cache: %v", err)
	}

	l, err := lock.Local()
	if err != nil {
		t.Fatalf("failed to init lock: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	first, err := initTestDistrbutedPodID(ctx, "test", "first", c, l)
	if err != nil {
		t.Fatalf("failed to init first pid: %v", err)
	}
	if first.ID() != 0 {
		t.Fatalf("expected first id to be 0 but was %d", first.ID())
	}

	second, err := initTestDistrbutedPodID(ctx, "test", "second", c, l)
	if err != nil {
		t.Fatalf("failed to init second pid: %v", err)
	}

	if second.ID() != 1 {
		t.Fatalf("expected second id to be 1 but was %d", second.ID())
	}

	third, err := initTestDistrbutedPodID(ctx, "test", "third", c, l)
	if err != nil {
		t.Fatalf("failed to init third pid: %v", err)
	}

	if third.ID() != 2 {
		t.Fatalf("expected third id to be 2 but was %d", third.ID())
	}

	if err := second.Release(ctx); err != nil {
		t.Fatalf("failed to release second pid: %v", err)
	}

	fourth, err := initTestDistrbutedPodID(ctx, "test", "fourth", c, l)
	if err != nil {
		t.Fatalf("failed to init third pid: %v", err)
	}

	if fourth.ID() != 1 {
		t.Fatalf("expected fourth id to be 1 but was %d", fourth.ID())
	}
}

func newTestDistributedPodID(prefix, ip string, cache cache.Cache, lock lock.DLocker) *DistributedPodID {
	return &DistributedPodID{
		prefix:            prefix,
		ip:                ip,
		cache:             cache,
		lock:              lock,
		renewWaitDuration: renewWaitDuration,
		renewSetDuration:  renewSetDuration,
		leaseTTLInSec:     cacheTTL,
		stopCh:            make(chan bool),
	}
}

func initTestDistrbutedPodID(ctx context.Context, prefix, ip string, cache cache.Cache, lock lock.DLocker) (*DistributedPodID, error) {
	pid := newTestDistributedPodID(prefix, ip, cache, lock)
	if err := pid.init(ctx); err != nil {
		return nil, err
	}
	return pid, nil
}
