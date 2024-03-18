package http_client

import (
	"context"
	"golang.org/x/net/context/ctxhttp"
	"net/http"
)

type httpClient struct {
	Client *http.Client
}

func (h *httpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return ctxhttp.Do(ctx, h.Client, req)
}

// NewHttpClient creates a new http client
// if client is nil, use http.DefaultClient
func NewHttpClient(client *http.Client) HttpClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &httpClient{
		Client: client,
	}
}
