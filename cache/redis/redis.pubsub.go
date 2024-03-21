package redis

import (
	"context"
	"github.com/dikiharyadi19/govegapunk/cache"
	"github.com/go-redis/redis/v8"
)

func (c *Cache) Publish(ctx context.Context, channel, message string) error {
	return c.client.Publish(ctx, channel, message).Err()
}

func (c *Cache) Subscribe(ctx context.Context, topic string) (cache.Subscriber, error) {
	redisPubSub := c.client.Subscribe(ctx, topic)
	pubSub := newCachePubSub(redisPubSub)

	_, err := redisPubSub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	return pubSub, nil
}

type cachePubSub struct {
	redisPubSub *redis.PubSub
}

func newCachePubSub(pubSub *redis.PubSub) *cachePubSub {
	return &cachePubSub{
		redisPubSub: pubSub,
	}
}

// Channel return string channel that will send the message payload
func (ps *cachePubSub) Channel() <-chan string {
	msgChan := make(chan string)

	go func() {
		defer close(msgChan)

		for redisMsg := range ps.redisPubSub.Channel() {
			if redisMsg != nil {
				msgChan <- redisMsg.Payload
			}
		}
	}()

	return msgChan
}

// ReceiveMessage return string that will send onto the channel
func (ps *cachePubSub) ReceiveMessage(ctx context.Context) (string, error) {
	msg, err := ps.redisPubSub.ReceiveMessage(ctx)
	if err != nil {
		return "", err
	}

	return msg.Payload, nil
}

// Close end redis pub-sub connection
func (ps *cachePubSub) Close() error {
	return ps.redisPubSub.Close()
}
