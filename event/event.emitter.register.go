package event

func RegisterSender(name string, fn SenderFactory) {
	senders[name] = fn
}

func RegisterWriter(name string, fn WriterFactory) {
	writers[name] = fn
}
