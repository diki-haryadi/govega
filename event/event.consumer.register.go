package event

var (
	listeners = map[string]ListenerFactory{
		"logger": EventLoggerListener,
	}
)

func RegisterListener(name string, fn ListenerFactory) {
	listeners[name] = fn
}
