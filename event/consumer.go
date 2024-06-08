package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/monitor"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	stop  uint32 = 0
	start uint32 = 1

	consumerGroupKey ctxKey = 1

	DefaultConsumerWorkers = 1
)

var (
	ErrConsumerStarted = errors.New("Consumer already started")

	errConfigNotFound = errors.New("config not found")

	listeners = map[string]ListenerFactory{
		"logger": EventLoggerListener,
	}
)

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

	WorkerPoolConfig map[string]interface{}

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

	EventConsumeMessage struct {
		Topic    string
		Key      string
		Metadata map[string]interface{}
		Data     []byte
	}

	eventConsumeMessageRead struct {
		Data     json.RawMessage        `json:"data" mapstructure:"data"`
		Metadata map[string]interface{} `json:"metadata,omitempty" mapstructure:"metadata"`
	}

	ListenerFactory func(ctx context.Context, config interface{}) (Listener, error)

	EventHandler    func(ctx context.Context, message *EventConsumeMessage) error
	EventMiddleware func(next EventHandler) EventHandler
	IteratorFunc    func(ctx context.Context) (ConsumeMessage, error)
	Job             func(ctx context.Context) error

	ctxKey int
)

// NewEventConsumeMessage return event consume message from byte data
func NewEventConsumeMessage(v []byte) (*EventConsumeMessage, error) {
	if v == nil {
		return &EventConsumeMessage{}, nil
	}

	var readmsg eventConsumeMessageRead
	if err := json.Unmarshal(v, &readmsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return &EventConsumeMessage{
		Metadata: readmsg.Metadata,
		Data:     readmsg.Data,
	}, nil
}

func RegisterListener(name string, fn ListenerFactory) {
	listeners[name] = fn
}

// GetConsumerGroupFromContext return consumer group from contet if any
func GetConsumerGroupFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(consumerGroupKey).(string); ok {
		return value
	}
	return ""
}

func (fn IteratorFunc) Next(ctx context.Context) (ConsumeMessage, error) {
	return fn(ctx)
}

func (c WorkerPoolConfig) getWorkers(topic, group string) int {
	config, err := c.parseTopicConfig(topic)
	if err != nil {
		if !errors.Is(err, errConfigNotFound) {
			log.WithError(err).Warnln("failed to parse topic config, fallback to default")
		}

		return c.getDefaultWorkers(topic, group)
	}

	if val, ok := config[group]; ok && val > 0 {
		return val
	}

	if val, ok := config[MetaDefault]; ok && val > 0 {
		return val
	}

	return c.getDefaultWorkers(topic, group)
}

func (c WorkerPoolConfig) parseTopicConfig(topic string) (map[string]int, error) {
	if val, ok := c[topic]; ok && val != nil {
		config := map[string]int{}

		if err := mapstructure.Decode(val, &config); err != nil {
			return map[string]int{}, fmt.Errorf("failed to decode config: %w", err)
		}

		return config, nil
	}

	return nil, errConfigNotFound
}

func (c WorkerPoolConfig) getDefaultWorkers(topic, group string) int {
	if val, ok := c[MetaDefault].(int); ok && val > 0 {
		return val
	}

	return DefaultConsumerWorkers
}

// NewConsumer create new instance of consumer
func NewConsumer(ctx context.Context, config *ConsumerConfig) (*Consumer, error) {

	if config == nil {
		return nil, errors.New("[event/consumer] missing config")
	}

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

			strategy, err := strategyFactory(ctx, config.ConsumeStrategy.Config)
			if err != nil {
				return nil,
					fmt.Errorf("[event/consumer] failed to init consumerstrategy [%s]: %w",
						config.ConsumeStrategy.Type, err)
			}

			consumer.consumeStrategy = strategy
		}
	}

	if config.Listener == nil {
		return nil, errors.New("[event/consumer] missing listener driver config")
	}

	factory, ok := listeners[config.Listener.Type]
	if !ok {
		return nil, fmt.Errorf("[event/consumer] unsupported listener driver: %s",
			config.Listener.Type)
	}

	listener, err := factory(ctx, config.Listener.Config)

	if err != nil {
		return nil, err
	}
	consumer.listener = listener

	return consumer, nil
}

// Use add middlewares to actual event handler before accessing the actual handler
// Please add your middlewares before calling subscribe or it may not work properly
func (c *Consumer) Use(middlewares ...EventMiddleware) {
	if len(c.listenerPools) > 0 {
		panic("[event/consumer] all middlewares should be added before subscribe")
	}
	c.middlewares = append(c.middlewares, middlewares...)
}

// Subscribe to a topic with specific group
// this should be call before Start the consumer
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

// WithConsumeStrategy, set consume strategy for this consumer
func (c *Consumer) WithConsumeStrategy(strategy ConsumeStrategy) {
	c.consumeStrategy = strategy
}

// Start activate the consumer and start receiving event
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

// Stop gracefully stop the consumer waiting for all workers to complete
// before exiting
func (c *Consumer) Stop() error {
	return c.StopContext(context.Background())
}

// StopContext gracefully stop the consumer or until the context timeout
func (c *Consumer) StopContext(ctx context.Context) error {
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

	var err error
	select {
	case <-ctx.Done():
		log.Errorln("[event/consumer] timeout waiting consumer to stop")
		err = ctx.Err()
	case <-c.shutdown:
	}

	c.stopch = nil
	c.shutdown = nil
	atomic.StoreUint32(&c.running, stop)

	log.Println("[event/consumer] stopped")
	return err
}

func (c *Consumer) isRunning() bool {
	return atomic.LoadUint32(&c.running) == start
}

type ListenerWorkerPool struct {
	workers         int
	iterator        Iterator
	handler         EventHandler
	topic           string
	group           string
	consumeStrategy ConsumeStrategy
	tracer          trace.Tracer
}

func (k *ListenerWorkerPool) run(stop <-chan bool, wg *sync.WaitGroup) {
	if closer, ok := k.iterator.(Closer); ok {
		defer func(closer Closer) {
			if err := closer.Close(); err != nil {
				log.WithError(err).
					Errorln("[listener/workerpool] failed to close iterator")
			}
		}(closer)
	}

	jobs := make(chan Job, k.workers)

	for i := 0; i < k.workers; i++ {
		go worker(jobs, stop, wg)
	}

	for {
		select {
		case <-stop:
			close(jobs)
			return
		default:
			jobs <- k.retrieveMessage
		}
	}

}

func (k *ListenerWorkerPool) retrieveMessage(ctx context.Context) error {
	ctx = context.WithValue(ctx, consumerGroupKey, k.group)

	message, err := k.iterator.Next(ctx)
	if err != nil {
		return fmt.Errorf("failed to get next item: %w", err)
	}

	if carrier, ok := message.(propagation.TextMapCarrier); ok {
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	ctx, span := k.tracer.Start(ctx, "listenerWorkerPool.consumeMessage",
		trace.WithAttributes(semconv.MessagingOperationProcess))
	defer span.End()

	start := time.Now()

	err = k.consumeStrategy(ctx, message, k.handler)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("failed to consume message topic [%s] group [%s]: %w",
			k.topic, k.group, err)
	}

	monitor.FeedConsumerMetrics(k.topic, k.group, getMetricStatusFromError(err),
		time.Since(start))

	return nil
}

func worker(jobs <-chan Job, stop <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stop:
			log.Println("[listener/worker] stop processing job")
			return
		case j := <-jobs:

			ctx, cancel := context.WithCancel(context.Background())
			resultCh := make(chan error, 1)

			go func(result chan<- error) {
				result <- j(ctx)
			}(resultCh)

			select {
			case <-stop:
				log.
					Println("[listener/worker] stopping, cancel context and wait job to complete")

				cancel()
				if err := <-resultCh; err != nil {
					logger := log.WithContext(ctx).WithError(err)

					if errors.Is(err, context.Canceled) {
						logger.Warnln("[listener/worker] cancel job error")
					} else {
						logger.Errorln("[listener/worker] unexpected error while canceling job")
					}
				}

				close(resultCh)
				return
			case err := <-resultCh:
				if err != nil {
					log.WithContext(ctx).WithError(err).
						Errorln("[listener/worker] failed to complete job")
				}

				cancel()
				close(resultCh)
			}
		}
	}
}

func getMetricStatusFromError(err error) string {
	if err == nil {
		return "OK"
	}

	return "ERROR"
}
