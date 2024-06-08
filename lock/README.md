# lock
Distributed Lock package, implementation
- local
- redis
- etcd
- zookeeper

## Usage

```
import (
    "github.com/diki-haryadi/govega/lock"
    _ "github.com/diki-haryadi/govega/lock/etcd"
    _ "github.com/diki-haryadi/govega/lock/redis"
    _ "github.com/diki-haryadi/govega/lock/zk"
)

url := "local://"
// url := "redis://localhost:6379/test"
// url := "redis://u1:p1@127.0.0.1:6379,u2:p2@127.0.0.2:6379,u3:p3@127.0.0.3:6379/test"
// url := "etcd://127.0.0.1:2379,127.0.0.2:2379,127.0.0.3:2379/test"
// url := "zk://127.0.0.1:2181,127.0.0.2:2181,127.0.0.3:2181/test"

dlock, _ := lock.New(url)

// Lock and wait
if err := dlock.Lock(ctx, id, 20); err != nil {
    return err
}

// Try lock and return immediately
if err := dlock.TryLock(ctx, id, 20); err != nil {
    continue
}

// Unlock
if err := dlock.Unlock(ctx, id); err != nil {
    return err
}

```