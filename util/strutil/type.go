package strutil

// StringerFunc utility type which transform function into a stringer
type StringerFunc func() string

func (fn StringerFunc) String() string {
	return fn()
}
