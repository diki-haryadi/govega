package taskworker

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type TaskWorker struct {
	workers        []*worker
	maxConcurrency int32
}

func NewTaskWorker(maxConcurrency uint8) *TaskWorker {
	return &TaskWorker{
		workers:        make([]*worker, 0),
		maxConcurrency: int32(maxConcurrency),
	}
}

func (t *TaskWorker) Register(work Work) {
	t.workers = append(t.workers, &worker{
		work: work,
	})
}

func (t *TaskWorker) Run(ctx context.Context) []Result {
	wg := sync.WaitGroup{}

	total := len(t.workers)
	temp := holder{
		res: make([]Result, 0),
	}

	wg.Add(total)

	i := 0
	for i < total {
		if temp.GetActiveWorker() < t.maxConcurrency {
			temp.Add()

			go func(index int) {
				defer func() {
					if r := recover(); r != nil {
						stack := debug.Stack()

						temp.Store(nil,
							fmt.Errorf("[taskworker] panic on running task, stacktrace: %s", string(stack)))
					}

					wg.Done()
				}()

				temp.Store(t.workers[index].work(ctx))
			}(i)
			i++
		}
	}

	wg.Wait()

	return temp.GetAllResult()
}

type worker struct {
	work Work
}

type Work func(ctx context.Context) (interface{}, error)
