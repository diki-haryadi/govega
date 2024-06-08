package event

import (
	"context"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	testListener struct {
		ch             chan ConsumeMessage
		customIterator Iterator
	}
	testConsumeMessage struct {
		em        *EventConsumeMessage
		committed bool
	}
)

func newTestListener() *testListener {
	return &testListener{
		ch: make(chan ConsumeMessage),
	}
}

func newTestConsumeMessage(em *EventConsumeMessage) *testConsumeMessage {
	return &testConsumeMessage{
		em:        em,
		committed: false,
	}
}

func (t *testConsumeMessage) GetEventConsumeMessage(ctx context.Context) (*EventConsumeMessage, error) {
	return t.em, nil
}

func (t *testConsumeMessage) Commit(ctx context.Context) error {
	t.committed = true
	return nil
}

func (t *testListener) factory(ctx context.Context, config interface{}) (Listener, error) {
	return t, nil
}

func (t *testListener) Listen(ctx context.Context, topic, group string) (Iterator, error) {
	if t.customIterator != nil {
		return t.customIterator, nil
	}

	return t, nil
}

func (t *testListener) sendMessage(msg ConsumeMessage) {
	go func() {
		t.ch <- msg
	}()
}

func (t *testListener) Next(ctx context.Context) (ConsumeMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-t.ch:
		return msg, nil
	}
}

func (t *testListener) Close() error {
	close(t.ch)
	return nil
}

func TestConsumerConsumeMessage(t *testing.T) {
	conf := &ConsumerConfig{
		Listener: &DriverConfig{
			Type: "TestConsumerConsumeMessage",
		},
	}

	testListener := newTestListener()
	RegisterListener("TestConsumerConsumeMessage", testListener.factory)

	ctx := context.Background()

	consumer, err := NewConsumer(ctx, conf)
	require.NoError(t, err)
	require.NotNil(t, consumer)

	var called uint32
	err = consumer.Subscribe(ctx, "test", "test", func(_ context.Context, _ *EventConsumeMessage) error {
		atomic.AddUint32(&called, 1)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, consumer.Start())
	testListener.sendMessage(newTestConsumeMessage(&EventConsumeMessage{}))
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, uint32(1), atomic.LoadUint32(&called))
	require.NoError(t, consumer.Stop())
}

func TestConsumerMiddleware(t *testing.T) {
	conf := &ConsumerConfig{
		Listener: &DriverConfig{
			Type: "TestConsumerMiddleware",
		},
	}

	testListener := newTestListener()
	RegisterListener("TestConsumerMiddleware", testListener.factory)

	ctx := context.Background()

	consumer, err := NewConsumer(ctx, conf)
	require.NoError(t, err)
	require.NotNil(t, consumer)

	var mwCalled int64
	mw := func(next EventHandler) EventHandler {
		return func(ctx context.Context, message *EventConsumeMessage) error {
			atomic.AddInt64(&mwCalled, time.Now().UnixNano())
			time.Sleep(50 * time.Millisecond)
			return next(ctx, message)
		}
	}
	consumer.Use(mw)

	var handlerCalled int64
	err = consumer.Subscribe(ctx, "test", "test", func(_ context.Context, _ *EventConsumeMessage) error {
		atomic.AddInt64(&handlerCalled, time.Now().UnixNano())
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, consumer.Start())
	testListener.sendMessage(newTestConsumeMessage(&EventConsumeMessage{}))
	time.Sleep(100 * time.Millisecond)

	mwc := atomic.LoadInt64(&mwCalled)
	hc := atomic.LoadInt64(&handlerCalled)

	assert.NotZero(t, mwc, "mwCalled should not equal to zero")
	assert.NotZero(t, hc, "handlerCalled should not equal to zero")

	assert.Greater(t, hc, mwc, "handlerCalled value should be bigger than mwCalled")
	require.NoError(t, consumer.Stop())
}

func TestConsumerGracefulStop(t *testing.T) {
	conf := &ConsumerConfig{
		Listener:         &DriverConfig{Type: "TestConsumerGracefulStop"},
		EventConfig:      &EventConfig{},
		WorkerPoolConfig: &WorkerPoolConfig{"default": 10},
	}

	testListener := newTestListener()

	var pools int32

	testListener.customIterator = IteratorFunc(func(ctx context.Context) (ConsumeMessage, error) {
		//context is cancel during stopping
		<-ctx.Done()
		r := rand.New(rand.NewSource(time.Now().Unix()))
		n := r.Intn(1500-100) + 100 // between 100 - 1500
		time.Sleep(time.Duration(n) * time.Millisecond)
		atomic.AddInt32(&pools, -1)

		return newTestConsumeMessage(&EventConsumeMessage{}), nil
	})

	RegisterListener("TestConsumerGracefulStop", testListener.factory)

	ctx := context.Background()

	consumer, err := NewConsumer(ctx, conf)
	require.NoError(t, err)
	require.NotNil(t, consumer)

	handler := func(ctx context.Context, _ *EventConsumeMessage) error {
		return nil
	}

	require.NoError(t, consumer.Subscribe(ctx, "test1", "test1", handler))
	require.NoError(t, consumer.Subscribe(ctx, "test2", "test2", handler))
	require.NoError(t, consumer.Subscribe(ctx, "test3", "test3", handler))

	atomic.StoreInt32(&pools, int32(conf.WorkerPoolConfig.getDefaultWorkers("", ""))*3)
	require.NoError(t, consumer.Start())

	time.Sleep(100 * time.Millisecond)
	require.NoError(t, consumer.Stop())
	assert.Equal(t, int32(0), atomic.LoadInt32(&pools))
}

func TestConsumerForceStop(t *testing.T) {
	conf := &ConsumerConfig{
		Listener:         &DriverConfig{Type: "TestConsumerForceStop"},
		EventConfig:      &EventConfig{},
		WorkerPoolConfig: &WorkerPoolConfig{"default": 10},
	}

	testListener := newTestListener()

	var pools int32

	testListener.customIterator = IteratorFunc(func(ctx context.Context) (ConsumeMessage, error) {
		//context is cancel during stopping
		<-ctx.Done()
		time.Sleep(1 * time.Minute)
		atomic.AddInt32(&pools, -1)

		return newTestConsumeMessage(&EventConsumeMessage{}), nil
	})

	RegisterListener("TestConsumerForceStop", testListener.factory)

	ctx := context.Background()

	consumer, err := NewConsumer(ctx, conf)
	require.NoError(t, err)
	require.NotNil(t, consumer)

	handler := func(ctx context.Context, _ *EventConsumeMessage) error {
		return nil
	}

	require.NoError(t, consumer.Subscribe(ctx, "test1", "test1", handler))
	require.NoError(t, consumer.Subscribe(ctx, "test2", "test2", handler))
	require.NoError(t, consumer.Subscribe(ctx, "test3", "test3", handler))

	totalPools := int32(conf.WorkerPoolConfig.getDefaultWorkers("", "") * 3)
	atomic.StoreInt32(&pools, totalPools)
	require.NoError(t, consumer.Start())

	time.Sleep(100 * time.Millisecond)
	stopCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = consumer.StopContext(stopCtx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, totalPools, atomic.LoadInt32(&pools))
}
