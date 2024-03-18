package event

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"bitbucket.org/rctiplus/vegapunk/log"
)

const (
	DefaultConsumerWorkers = 1
)

type WorkerPoolConfig map[string]interface{}

type ListenerWorkerPool struct {
	workers         int
	iterator        Iterator
	handler         EventHandler
	topic           string
	group           string
	consumeStrategy ConsumeStrategy
	tracer          trace.Tracer
}

func (c WorkerPoolConfig) getWorkers(topic, group string) int {
	if m, ok := c[topic].(map[string]int); ok {
		if val, ok := m[group]; ok && val > 0 {
			return val
		}

		if val, ok := m[MetaDefault]; ok && val > 0 {
			return val
		}
	}

	return c.getDefaultWorkers(topic, group)
}

func (c WorkerPoolConfig) getDefaultWorkers(topic, group string) int {
	if val, ok := c[MetaDefault].(int); ok && val > 0 {
		return val
	}

	return DefaultConsumerWorkers
}

func (k *ListenerWorkerPool) run(stop <-chan bool, wg *sync.WaitGroup) {
	if closer, ok := k.iterator.(Closer); ok {
		defer func(closer Closer) {
			if err := closer.Close(); err != nil {
				log.WithError(err).
					Errorln("[event/consumer] failed to close iterator")
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
		return fmt.Errorf("[event/consumer] failed to get next item: %w", err)
	}

	if carrier, ok := message.(propagation.TextMapCarrier); ok {
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	ctx, span := k.tracer.Start(ctx, "listenerWorkerPool.consumeMessage",
		trace.WithAttributes(semconv.MessagingOperationProcess))
	defer span.End()

	if err := k.consumeStrategy(ctx, message, k.handler); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("[event/consumer] failed to consume message topic [%s] group [%s]: %w",
			k.topic, k.group, err)
	}

	return nil
}

func worker(jobs <-chan Job, stop <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stop:
			log.Println("[event/consumer] stop processing job")
			return
		case j := <-jobs:

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan bool, 1)

			go func(done chan<- bool) {
				if err := j(ctx); err != nil {
					//NOTE: add backoff?
					log.WithContext(ctx).WithError(err).
						Errorln("[event/consumer] failed to complete job")
				}
				close(done)
			}(done)

			select {
			case <-stop:
				log.Println("[event/consumer] stopping, cancel context and wait job to complete")
				cancel()
				<-done
				return
			case <-done:
				cancel()
			}
		}
	}
}
