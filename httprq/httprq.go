package httprq

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	Request struct {
		URL                   string
		Method                string
		Body                  io.Reader
		Params                map[string]string
		Headers               map[string]string
		Context               context.Context
		Timeout               int
		TimeoutHystrix        int
		MaxConcurrentRequests int
		ErrorPercentThreshold int
		ErrNotSuccess         bool
	}

	RequestBuilder struct {
		request        Request
		requestManager RequestManager
		requestRetry   RequestRetry
	}
)

func Do(method, url string) *RequestBuilder {
	return &RequestBuilder{
		request: Request{
			URL:     url,
			Method:  method,
			Headers: map[string]string{},
		},
		requestManager: requestManager,
		requestRetry:   &RequestRetryWhenTimeout{},
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

func (rb *RequestBuilder) WithTimeout(timeout int) *RequestBuilder {
	rb.request.Timeout = timeout
	return rb
}

func (rb *RequestBuilder) WithTimeoutHystrix(timeout int, maxConcurrentRequests int, errorPercentThreshold int) *RequestBuilder {
	rb.request.TimeoutHystrix = timeout
	rb.request.MaxConcurrentRequests = maxConcurrentRequests
	rb.request.ErrorPercentThreshold = errorPercentThreshold
	return rb
}

func (rb *RequestBuilder) WithRetryStrategyWhenTimout(attempt int) *RequestBuilder {
	rb.requestRetry = &RequestRetryWhenTimeout{
		RequestNoRetry: RequestNoRetry{},
		RequestConfig: RequestConfig{
			NumRetry: attempt,
			Delay:    BackOffDelay(time.Second),
		},
	}
	return rb
}

func (rb *RequestBuilder) WithRetryStrategyAllErrors(attempt int) *RequestBuilder {
	rb.requestRetry = &RequestRetryAllErrors{
		RequestNoRetry: RequestNoRetry{},
		RequestConfig: RequestConfig{
			NumRetry: attempt,
			Delay:    BackOffDelay(time.Second),
		},
	}
	return rb
}

func (rb *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	rb.request.Headers[key] = value
	return rb
}

func (rb *RequestBuilder) AddHeaders(headers map[string]string) *RequestBuilder {
	if rb.request.Params == nil {
		rb.request.Params = make(map[string]string)
	}

	for k, v := range headers {
		rb.request.Headers[k] = v
	}
	return rb
}

func (rb *RequestBuilder) AddQueryParam(key, value string) *RequestBuilder {
	rb.request.Headers[key] = value
	return rb
}

func (rb *RequestBuilder) AddQueryParams(params map[string]string) *RequestBuilder {
	if rb.request.Params == nil {
		rb.request.Params = make(map[string]string)
	}

	for k, v := range params {
		rb.request.Params[k] = v
	}
	return rb
}

func (rb *RequestBuilder) SetBasicAuth(username string, password string) *RequestBuilder {
	rb.request.Headers["authorization"] = fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", username, password)),
		))
	return rb
}

func (rb *RequestBuilder) SetErrNotSuccess(errNotSuccess bool) *RequestBuilder {
	rb.request.ErrNotSuccess = errNotSuccess
	return rb
}

func (rb *RequestBuilder) Execute() (response *Response) {
	return rb.requestRetry.DoRequest(rb.requestManager, rb.request)
}

func (rb *RequestBuilder) ExecuteWithoutHystrix() *Response {
	return rb.requestRetry.DoRequest(rb.requestManager, rb.request)
}
