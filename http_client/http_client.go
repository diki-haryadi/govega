package http_client

import (
	"context"
	"net/http"
)

type HttpClient interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
