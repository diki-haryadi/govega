package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	URL               string
	Method            string
	Body              io.Reader
	Headers           map[string]string
	Context           context.Context
	Timeout           int
	OnRequestStart    OnRequestStartFunc
	OnRequestFinished OnRequestFinishedFunc
}

type OnRequestStartFunc func(*http.Request) error
type OnRequestFinishedFunc func(*Result) error

type RequestBuilder struct {
	request       Request
	client        Client
	retryStrategy RetryStrategy
}

func Do(method, url string) *RequestBuilder {
	return &RequestBuilder{
		request: Request{
			URL:     url,
			Method:  method,
			Headers: map[string]string{},
		},
		retryStrategy: &NoRetry{},
		client:        defaultClient,
	}
}

func Post(url string) *RequestBuilder {
	return Do(http.MethodPost, url)
}

func Get(url string) *RequestBuilder {
	return Do(http.MethodGet, url)
}

func Put(url string) *RequestBuilder {
	return Do(http.MethodPut, url)
}

func Delete(url string) *RequestBuilder {
	return Do(http.MethodDelete, url)
}

func (rb *RequestBuilder) WithBody(body io.Reader) *RequestBuilder {
	rb.request.Body = body
	return rb
}

func (rb *RequestBuilder) WithContext(ctx context.Context) *RequestBuilder {
	rb.request.Context = ctx
	return rb
}

func (rb *RequestBuilder) WithClient(c Client) *RequestBuilder {
	rb.client = c
	return rb
}

func (rb *RequestBuilder) WithRetryStrategy(rs RetryStrategy) *RequestBuilder {
	rb.retryStrategy = rs
	return rb
}

func (rb *RequestBuilder) WithTimeout(timeout int) *RequestBuilder {
	rb.request.Timeout = timeout
	return rb
}

func (rb *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	rb.request.Headers[key] = value
	return rb
}

func (rb *RequestBuilder) AddHeaders(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.request.Headers[k] = v
	}
	return rb
}

func (rb *RequestBuilder) OnRequestStart(f OnRequestStartFunc) *RequestBuilder {
	rb.request.OnRequestStart = f
	return rb
}

func (rb *RequestBuilder) OnRequestFinished(f OnRequestFinishedFunc) *RequestBuilder {
	rb.request.OnRequestFinished = f
	return rb
}

func (rb *RequestBuilder) Execute() *Result {
	return rb.retryStrategy.DoRequest(rb.client, rb.request)
}

func (rb *RequestBuilder) SetBasicAuth(username string, password string) *RequestBuilder {
	rb.request.Headers["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
	return rb
}
