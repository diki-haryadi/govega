package cache

const (
	ErrNotFound         = CacheError("[cache] not found")
	ErrUnsuportedSchema = CacheError("[cache] unsupported scheme")
)

type CacheError string

func (e CacheError) Error() string {
	return string(e)
}
