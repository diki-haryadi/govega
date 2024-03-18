package cache

var (
	cacheImpl = make(map[string]InitFunc)
)

func Register(schema string, fn InitFunc) {
	cacheImpl[schema] = fn
}
