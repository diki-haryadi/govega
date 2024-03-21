package redis

import "github.com/dikiharyadi19/govegapunk/cache"

const (
	schemaDefault = "redis"
	schema        = "redis"
)

func init() {
	cache.Register(schema, NewCache)
}
