package kafka

import "bitbucket.org/rctiplus/vegapunk/event"

func init() {
	event.RegisterSender("kafka", NewKafkaSender)
	event.RegisterListener("kafka", NewKafkaListener)
}
