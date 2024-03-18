package http_client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type Counters struct {
	mu      sync.Mutex
	counter int
}

func (c *Counters) inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter++
}

func NewMockServer(timeoutMs int, C *Counters) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(timeoutMs) * time.Millisecond)

		if C != nil {
			C.inc()
		}

		w.WriteHeader(200)
		_, _ = w.Write([]byte("OK"))
	}))
	return s
}

func TestHttpClient_Do(t *testing.T) {
	s := NewMockServer(0, nil)
	defer s.Close()

	t.Run("it should make http call", func(t *testing.T) {
		client := NewHttpClient(nil)

		req, _ := http.NewRequest("GET", s.URL, nil)
		resp, err := client.Do(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	})
}

func TestHttpClientWithTimeout_Do(t *testing.T) {
	t.Run("it should return success when under timeout", func(t *testing.T) {
		s := NewMockServer(0, nil)
		defer s.Close()

		client := NewHttpClientWithTimeout(NewHttpClient(nil), 100, 100, 10)

		req, _ := http.NewRequest("GET", s.URL, nil)
		resp, err := client.Do(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	})

	t.Run("it should return error when over timeout", func(t *testing.T) {
		s := NewMockServer(200, nil)
		defer s.Close()

		client := NewHttpClientWithTimeout(NewHttpClient(nil), 50, 100, 10)

		req, _ := http.NewRequest("GET", s.URL, nil)
		_, err := client.Do(context.Background(), req)
		if err == nil {
			t.Errorf("expected error")
		}
	})
}

func TestHttpWithSingleFlight_Do(t *testing.T) {
	t.Run("it should handle multiple parallel requests", func(t *testing.T) {
		counter := Counters{}
		s := NewMockServer(100, &counter)
		defer s.Close()

		client := NewHttpWithSingleFlight(NewHttpClient(nil))

		var wg sync.WaitGroup
		req, _ := http.NewRequest("GET", s.URL, nil)

		calls := 5
		wg.Add(calls)
		for i := 0; i < calls; i++ {
			go func(req *http.Request) {
				_, _ = client.Do(context.Background(), req)
				wg.Done()
			}(req)
		}
		wg.Wait()

		if counter.counter != 1 {
			t.Errorf("expected 1 call, got %d", counter.counter)
		}
	})
}
