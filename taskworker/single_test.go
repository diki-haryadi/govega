package taskworker

import (
	"context"
	"testing"
)

func TestSingleWorker(t *testing.T) {
	max := 200
	tasker := NewSingleTaskWorker(context.Background(), 10, func(ctx context.Context, input interface{}) (interface{}, error) {
		res, err := DummyProcessSuccess()
		return res, err
	}, max)

	for i := 0; i < max; i++ {
		tasker.Do(i)
	}

	result := tasker.Results()
	if len(result) != max {
		t.Fatalf("failed to finish all task")
	}

}
