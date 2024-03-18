package redis

import "bitbucket.org/rctiplus/vegapunk/cache"

const (
	schemaDefault = "redis"
	schema        = "redis"
)

func init() {
	cache.Register(schema, NewCache)
}
