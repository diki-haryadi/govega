package redis

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dikiharyadi19/govegapunk/lock"
	"github.com/dikiharyadi19/govegapunk/lock/redis/redlock"
)

const (
	schema = "redis"
)

type LockManager struct {
	manager *redlock.RedLock
	prefix  string
}

func init() {
	lock.Register(schema, New)
}

func New(urls []*url.URL) (lock.DLocker, error) {
	hs := make([]string, 0)
	for _, v := range urls {
		host := "tcp://"

		pass, ok := v.User.Password()
		if ok {
			host += fmt.Sprintf("%s:%s@", v.User.Username(), pass)
		}

		host += v.Host
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
