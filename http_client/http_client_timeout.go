package http_client

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"net/http"
)

type httpClientWithTimeout struct {
	httpClient     HttpClient
	hystrixCommand string
}

func (h *httpClientWithTimeout) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	err = hystrix.Do(h.hystrixCommand, func() error {
		resp, err = h.httpClient.Do(ctx, req)
		return err
	}, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// NewHttpClientWithTimeout wraps http client with hystrix timeout
func NewHttpClientWithTimeout(client HttpClient, timeout int, maxConcurrentRequests int, errorPercentThreshold int) HttpClient {
	command := "general"
	hystrix.Configure(map[string]hystrix.CommandConfig{
		command: {
			Timeout:               timeout,
			MaxConcurrentRequests: maxConcurrentRequests,
			ErrorPercentThreshold: errorPercentThreshold,
		},
	})
	return &httpClientWithTimeout{
		httpClient:     client,
		hystrixCommand: command,
	}
}
