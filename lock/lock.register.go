package lock

var (
	lockerImpl = make(map[string]InitFunc)
)

func Register(schema string, fn InitFunc) {
	lockerImpl[schema] = fn
}
