package http_client

import (
	"context"
	"errors"
	"golang.org/x/sync/singleflight"
	"net/http"
)

type httpWithSingleFlight struct {
	client HttpClient
	g      singleflight.Group
}

func (h *httpWithSingleFlight) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	val, err, _ := h.g.Do(req.URL.String(), func() (interface{}, error) {
		return h.client.Do(ctx, req)
	})
	if err != nil {
		return nil, err
	}

	resp, ok := val.(*http.Response)
	if !ok {
		return nil, errors.New("failed to get response")
	}

	return resp, nil
}

// NewHttpWithSingleFlight wraps http client with singleflight
func NewHttpWithSingleFlight(client HttpClient) HttpClient {
	return &httpWithSingleFlight{
		client: client,
		g:      singleflight.Group{},
	}
}
