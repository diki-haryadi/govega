package embed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/diki-haryadi/govega/cache"
)

const schema = "embed"

func init() {
	cache.Register(schema, NewBadgerCache)
}

type BadgerCache struct {
	db *badger.DB
}

func NewBadgerCache(url *url.URL) (cache.Cache, error) {

	opt := badger.DefaultOptions(url.Host + url.Path)
	if url.Host == "mem" {
		opt = badger.DefaultOptions("").WithInMemory(true)
	}

	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return &BadgerCache{
		db: db,
	}, nil
}

func (b *BadgerCache) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	return b.db.Update(func(txn *badger.Txn) error {

		var bin []byte
		switch v := value.(type) {
		case string:
			bin = []byte(v)
		case []byte:
			bin = v
		case bool:
			if v {
				bin = []byte("1")
			} else {
				bin = []byte("0")
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			bin = []byte(fmt.Sprintf("%v", v))
		default:
			var err error
			bin, err = json.Marshal(value)
			if err != nil {
				return err
			}
		}

		e := badger.NewEntry([]byte(key), bin)
		if expiration > 0 {
			e = e.WithTTL(time.Second * time.Duration(expiration))
		}
		return txn.SetEntry(e)
	})
}

func (c *BadgerCache) Increment(ctx context.Context, key string, expiration int) (int64, error) {
	return 0, cache.NotSupported
}

func (b *BadgerCache) Get(ctx context.Context, key string) ([]byte, error) {
	var out []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			out = val
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (b *BadgerCache) GetObject(ctx context.Context, key string, doc interface{}) error {
	return b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, doc)
		})
	})
}

func (b *BadgerCache) GetString(ctx context.Context, key string) (string, error) {
	var out string
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			out = string(val)
			return nil
		})
	})
	if err != nil {
		return "", err
	}

	return out, nil
}

func (b *BadgerCache) GetInt(ctx context.Context, key string) (int64, error) {
	var out int64
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			i, err := strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				return err
			}
			out = i
			return nil
		})
	})
	if err != nil {
		return 0, err
	}

	return out, nil
}

func (b *BadgerCache) GetFloat(ctx context.Context, key string) (float64, error) {
	var out float64
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			f, err := strconv.ParseFloat(string(val), 64)
			if err != nil {
				return err
			}
			out = f
			return nil
		})
	})
	if err != nil {
		return 0, err
	}

	return out, nil
}

func (b *BadgerCache) Exist(ctx context.Context, key string) bool {
	err := b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})

	return err == nil
}

func (b *BadgerCache) Delete(ctx context.Context, key string, opts ...cache.DeleteOptions) error {
	deleteCache := &cache.DeleteCache{}
	for _, opt := range opts {
		opt(deleteCache)
	}

	if deleteCache.Pattern != "" {
		return b.deletePattern(ctx, deleteCache.Pattern)
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (b *BadgerCache) deletePattern(ctx context.Context, pattern string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(pattern)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			txn.Delete(k)
		}
		return nil
	})
}

func (b *BadgerCache) GetKeys(ctx context.Context, pattern string) []string {
	out := make([]string, 0)
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(pattern)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			out = append(out, string(k))
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return out
}

func (b *BadgerCache) RemainingTime(ctx context.Context, key string) int {
	rem := uint64(0)
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		rem = item.ExpiresAt() - uint64(time.Now().Unix())
		return nil
	})

	if err != nil {
		return -1
	}
	return int(rem)
}

func (b *BadgerCache) Close() error {
	return b.db.Close()
}
