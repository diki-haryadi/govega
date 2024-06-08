package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func init() {
	SetDefaultClient(&DefaultClient{
		Client:  &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		Timeout: 1,
	})
}

var defaultClient Client

func SetDefaultClient(c Client) {
	defaultClient = c
}

func GetDefaultClient() Client {
	return defaultClient
}
