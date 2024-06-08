package util

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diki-haryadi/govega/cache"
	"github.com/diki-haryadi/govega/lock"
	"github.com/diki-haryadi/govega/log"
)

var (
	cacheTTL          = 60 * 60 * 24 //1 day
	initTimeout       = 10 * time.Second
	renewWaitDuration = 23 * time.Hour
	renewSetDuration  = 30 * time.Second

	ErrPodIDExhausted       = errors.New("all available pod id has been acquired")
	ErrMachineIPUnavailable = errors.New("unable to determine machine ip address")

	errPodIDAcquired = errors.New("pod id has been acquired")
)

type DistributedPodID struct {
	prefix            string
	cache             cache.Cache
	lock              lock.DLocker
	ip                string
	acquiredID        int64
	key               string
	renewWaitDuration time.Duration
	renewSetDuration  time.Duration
	leaseTTLInSec     int

	mu      sync.Mutex
	stopped uint32
	stopCh  chan bool
}

func NewDistributedPodID(ctx context.Context, prefix string, cache cache.Cache,
	lock lock.DLocker) (*DistributedPodID, error) {
	ip, err := getIP()
	if err != nil {
		return nil, err
	}

	pid := DistributedPodID{
		prefix:            prefix,
		cache:             cache,
		lock:              lock,
		ip:                ip,
		renewWaitDuration: renewWaitDuration,
		renewSetDuration:  renewSetDuration,
		leaseTTLInSec:     cacheTTL,
		mu:                sync.Mutex{},
		stopped:           0,
		stopCh:            make(chan bool),
	}

	if err := pid.init(ctx); err != nil {
		return nil, err
	}

	return &pid, nil
}

// ID return the acquired pod id.
func (d *DistributedPodID) ID() int64 {
	return d.acquiredID
}

// Release acquired pod id from cache.
func (d *DistributedPodID) Release(ctx context.Context) error {
	if atomic.LoadUint32(&d.stopped) == 1 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if atomic.LoadUint32(&d.stopped) == 1 {
		return nil
	}

	d.stopCh <- true
	return d.cache.Delete(ctx, d.key)
}

func (d *DistributedPodID) init(ctx context.Context) error {
	for i := int64(0); i < maxNode; i++ {
		idKey := d.getIDKey(i)

		value, err := d.getDataFromCache(ctx, idKey)
		if err != nil {
			return err
		}

		if value != "" && value != d.ip {
			continue
		}

		if err := d.acquireID(ctx, i); err != nil {
			if errors.Is(err, errPodIDAcquired) ||
				errors.Is(err, lock.ErrResourceLocked) {
				continue
			}
			return fmt.Errorf("failed to acquire id: %w", err)
		}

		d.acquiredID = i
		d.key = idKey
		d.startRenewWorker()
		return nil
	}

	if d.key == "" {
		return ErrPodIDExhausted
	}

	return nil
}

func (d *DistributedPodID) getDataFromCache(ctx context.Context, key string) (string, error) {
	value, err := d.cache.GetString(ctx, key)
	if err != nil && !errors.Is(err, cache.NotFound) {
		return "", fmt.Errorf("failed to get cache data: %w", err)
	}

	return value, nil
}

func (d *DistributedPodID) acquireID(ctx context.Context, num int64) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(initTimeout)

		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	lockKey := d.getLockKey(num)
	lockTTL := int(time.Until(deadline).Seconds())

	if err := d.lock.TryLock(ctx, lockKey, lockTTL); err != nil {
		return fmt.Errorf("failed to retrieve lock: %w", err)
	}
	defer d.unlock(ctx, lockKey)

	idKey := d.getIDKey(num)

	value, err := d.getDataFromCache(ctx, idKey)
	if err != nil {
		return err
	}

	if value != "" && value != d.ip {
		return errPodIDAcquired
	}

	if err := d.cache.Set(ctx, idKey, d.ip, d.leaseTTLInSec); err != nil {
		return fmt.Errorf("failed to set cache data: %w", err)
	}
	return nil
}

func (d *DistributedPodID) getIDKey(num int64) string {
	return fmt.Sprintf("%s:distributed_pod:id:%d", d.prefix, num)
}

func (d *DistributedPodID) getLockKey(num int64) string {
	return fmt.Sprintf("%s:distributed_pod:lock:%d", d.prefix, num)
}

func (d *DistributedPodID) startRenewWorker() {
	go func() {
		tick := time.NewTicker(d.renewWaitDuration)

		for {
			select {
			case <-tick.C:
				if err := d.renewLease(); err != nil {
					log.WithError(err).Errorln("failed to refresh pod id")
				}
			case <-d.stopCh:
				atomic.StoreUint32(&d.stopped, 1)
				return
			}
		}
	}()
}

func (d *DistributedPodID) renewLease() error {
	ctx, cancel := context.WithTimeout(context.Background(), d.renewSetDuration)
	defer cancel()

	return d.cache.Set(ctx, d.key, d.ip, d.leaseTTLInSec)
}

func (d *DistributedPodID) unlock(ctx context.Context, key string) {
	var isContextDone bool
	select {
	case <-ctx.Done():
		isContextDone = true
	default:
	}

	if isContextDone {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}

	if err := d.lock.Unlock(ctx, key); err != nil {
		log.WithError(err).Errorln("[util/distributedpodid] failed to unlock")
	}
}

func getIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || // interface down
			iface.Flags&net.FlagLoopback != 0 { // loopback interface
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.WithError(err).Errorln("failed on retrieving interface address")
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}

	return "", ErrMachineIPUnavailable
}
