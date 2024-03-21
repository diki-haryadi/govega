package kafka

import "github.com/dikiharyadi19/govegapunk/event"

func init() {
	event.RegisterSender("kafka", NewKafkaSender)
	event.RegisterListener("kafka", NewKafkaListener)
}
