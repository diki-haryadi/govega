package main

import (
	"context"
	"fmt"
	"github.com/dikiharyadi19/govegapunk/cache/redis"
	"net/url"
)

var client redis.Cache

func main() {
	redisCreds := url.URL{
		Host: "localhost:6379",
		User: url.UserPassword("", "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81"),
	}
	client, err := redis.NewCache(&redisCreds)
	if err != nil {
		fmt.Println(err)
	}

	// Set Data to redis
	err = client.Set(context.Background(), "token", "data", 10)
	if err != nil {
		fmt.Println(err)
	}

	// Get Data from redis
	data, err := client.Get(context.Background(), "token")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(data))
	redisPublisher()
	redisSubscriber()
}

func redisPublisher() {
	if err := client.Publish(context.Background(), "send-user-data", "data"); err != nil {
		panic(err)
	}
}

func redisSubscriber() {
	subscriber, _ := client.Subscribe(context.Background(), "send-user-data")
	for {
		msg, err := subscriber.ReceiveMessage(context.Background())
		if err != nil {
			panic(err)
		}

		fmt.Println(msg)
	}
}
