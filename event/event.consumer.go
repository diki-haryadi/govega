package event

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"

	"bitbucket.org/rctiplus/vegapunk/log"
)

const (
	stop  uint32 = 0
	start uint32 = 1

	consumerGroupKey ctxKey = 1
)

type ctxKey int

type Job func(ctx context.Context) error
type ListenerFactory func(ctx context.Context, config interface{}) (Listener, error)
type EventHandler func(ctx context.Context, message *EventConsumeMessage) error
type EventMiddleware func(next EventHandler) EventHandler

type (
	Listener interface {
		Listen(ctx context.Context, topic, group string) (Iterator, error)
	}

	ConsumeMessage interface {
		GetEventConsumeMessage(ctx context.Context) (*EventConsumeMessage, error)
		Commit(ctx context.Context) error
	}

	Iterator interface {
		Next(ctx context.Context) (ConsumeMessage, error)
	}

	Closer interface {
		Close() error
	}

	Consumer struct {
		listener         Listener
		listenerPools    []ListenerWorkerPool
		middlewares      []EventMiddleware
		eventConfig      *EventConfig
		workerPoolConfig *WorkerPoolConfig
		consumeStrategy  ConsumeStrategy
		running          uint32
		lock             sync.Mutex
		stopch           chan bool
		shutdown         chan bool
	}

	ConsumerConfig struct {
		Listener         *DriverConfig     `json:"listener" mapstructure:"listener"`
		EventConfig      *EventConfig      `json:"event_config" mapstructure:"event_config"`
		WorkerPoolConfig *WorkerPoolConfig `json:"worker_pool_config" mapstructure:"worker_pool_config"`
		ConsumeStrategy  *DriverConfig     `json:"consume_strategy" mapstructure:"consume_strategy"`
	}
)

func NewConsumer(ctx context.Context, config *ConsumerConfig) (*Consumer, error) {

	if config == nil {
		return nil, errors.New("[event/consumer] missing config")
	}

	// default
	consumer := &Consumer{
		listenerPools:    make([]ListenerWorkerPool, 0),
		middlewares:      make([]EventMiddleware, 0),
		eventConfig:      &EventConfig{},
		workerPoolConfig: &WorkerPoolConfig{},
		consumeStrategy:  CommitOnSuccessStrategy,
		running:          stop,
		lock:             sync.Mutex{},
		stopch:           make(chan bool, 2),
	}

	if config != nil {
		if config.EventConfig != nil {
			consumer.eventConfig = config.EventConfig
		}

		if config.WorkerPoolConfig != nil {
			consumer.workerPoolConfig = config.WorkerPoolConfig
		}

		if config.ConsumeStrategy != nil {
			strategyFactory, ok := consumeStrategy[config.ConsumeStrategy.Type]
			if !ok {
				return nil,
					fmt.Errorf("[event/consumer] invalid consume strategy [%s]", config.ConsumeStrategy)
			}

			strategy, err := strategyFactory(ctx, config.ConsumeStrategy.Type)
			if err != nil {
				return nil, fmt.Errorf("[event/consumer] failed to init consumerstrategy [%s]: %w",
					config.ConsumeStrategy.Type, err)
			}

			consumer.consumeStrategy = strategy
		}
	}

	if config.Listener == nil {
		return nil, errors.New("[event/consumer] missing listener driver config")
	}

	listenerFactory, ok := listeners[config.Listener.Type]
	if !ok {
		return nil, fmt.Errorf("[event/consumer] unsupported listener driver: %s",
			config.Listener.Type)
	}

	listener, err := listenerFactory(ctx, config.Listener.Config)
	if err != nil {
		return nil, err
	}

	consumer.listener = listener
	return consumer, nil
}

func (c *Consumer) isRunning() bool {
	return atomic.LoadUint32(&c.running) == start
}

//WithConsumeStrategy, set consume strategy for this consumer
func (c *Consumer) WithConsumeStrategy(strategy ConsumeStrategy) {
	c.consumeStrategy = strategy
}

//Use add middlewares to actual event handler before accessing the actual handler
//Please add your middlewares before calling subscribe or it may not work properly
func (c *Consumer) Use(middlewares ...EventMiddleware) {
	if len(c.listenerPools) > 0 {
		panic("[event/consumer] all middlewares should be added before subscribe")
	}
	c.middlewares = append(c.middlewares, middlewares...)
}

//Subscribe to a topic with specific group
//this should be call before Start the consumer
func (c *Consumer) Subscribe(ctx context.Context, topic, group string,
	handler EventHandler) error {
	if c.isRunning() {
		return ErrConsumerStarted
	}

	topic = c.eventConfig.getTopic(topic)
	group = c.eventConfig.getGroup(group)

	iterator, err := c.listener.Listen(ctx, topic, group)
	if err != nil {
		return fmt.
			Errorf("[event/consumer] failed to get listener for topic: %w", err)
	}

	if len(c.middlewares) > 0 {
		for i := len(c.middlewares) - 1; i >= 0; i-- {
			handler = c.middlewares[i](handler)
		}
	}

	c.listenerPools = append(c.listenerPools, ListenerWorkerPool{
		workers:         c.workerPoolConfig.getWorkers(topic, group),
		iterator:        iterator,
		handler:         handler,
		consumeStrategy: c.consumeStrategy,
		topic:           topic,
		group:           group,
		tracer:          otel.Tracer("event/consumer"),
	})

	return nil
}

//Start activate the consumer and start receiving event
func (c *Consumer) Start() error {
	if c.isRunning() {
		return ErrConsumerStarted
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// double lock check
	// in case it started at the same time on different thread
	if c.isRunning() {
		return ErrConsumerStarted
	}

	c.stopch = make(chan bool)
	c.shutdown = make(chan bool)
	wg := sync.WaitGroup{}

	for i := range c.listenerPools {
		//NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		pool := c.listenerPools[i]

		//must be assign here, else race will be detected
		wg.Add(pool.workers)
		go pool.run(c.stopch, &wg)
	}

	go func() {
		wg.Wait()
		close(c.shutdown)
	}()

	atomic.StoreUint32(&c.running, start)

	return nil
}

//Stop gracefully stop the consumer waiting for all workers to complete
//before exiting
func (c *Consumer) Stop() error {
	if !c.isRunning() {
		return nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// double lock check
	// in case it stop at the same time on different thread
	if !c.isRunning() {
		return nil
	}

	log.Println("[event/consumer] stopping consumer")
	close(c.stopch)

	log.Println("[event/consumer] waiting for all workers to stop")
	<-c.shutdown

	c.stopch = nil
	c.shutdown = nil
	atomic.StoreUint32(&c.running, stop)

	log.Println("[event/consumer] stopped")
	return nil
}
