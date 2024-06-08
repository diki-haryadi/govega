package taskworker

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type SingleWork func(ctx context.Context, input interface{}) (interface{}, error)

type SingleTaskWorker struct {
	work   SingleWork
	wg     *sync.WaitGroup
	jobs   chan interface{}
	holder *holder
}

func NewSingleTaskWorker(ctx context.Context, maxConcurrency uint8, work SingleWork, total int) *SingleTaskWorker {
	wg := sync.WaitGroup{}
	wg.Add(total)

	jobs := make(chan interface{})

	t := &SingleTaskWorker{
		work:   work,
		wg:     &wg,
		jobs:   jobs,
		holder: &holder{},
	}

	for i := 0; i < int(maxConcurrency); i++ {
		go t.worker(ctx)
	}

	return t
}

func (t *SingleTaskWorker) Do(input interface{}) {
	t.jobs <- input
}

func (t *SingleTaskWorker) worker(ctx context.Context) {
	for job := range t.jobs {
		t.run(ctx, job)
	}
}

func (t *SingleTaskWorker) run(ctx context.Context, job interface{}) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			t.holder.Store(nil,
				fmt.Errorf("[singletaskworker] panic on running task, stacktrace: %s", string(stack)))
		}

		t.wg.Done()
	}()

	t.holder.Store(t.work(ctx, job))
}

func (t *SingleTaskWorker) Results() []Result {
	t.wg.Wait()
	close(t.jobs)

	return t.holder.res
}
