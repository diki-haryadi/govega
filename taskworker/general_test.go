package taskworker

import (
	"context"
	"errors"
	"testing"
)

func TestWorker(t *testing.T) {
	tw := NewTaskWorker(10)

	for i := 0; i < 1000; i++ {
		tw.Register(func(ctx context.Context) (interface{}, error) {
			resp, err := DummyProcessSuccess()
			return resp, err
		})
	}

	for i := 0; i < 1000; i++ {
		tw.Register(func(ctx context.Context) (interface{}, error) {
			resp, err := DummyProcessFail()
			return resp, err
		})
	}

	result := tw.Run(context.Background())

	if len(result) != 2000 {
		t.Fatalf("failed to finish all task")
	}
}

func DummyProcessSuccess() (string, error) {
	return "success", nil
}

func DummyProcessFail() (string, error) {
	return "Fail", errors.New("fail")
}
