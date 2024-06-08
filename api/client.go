package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/diki-haryadi/govega/custerr"
	"github.com/diki-haryadi/govega/log"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrEmptyResponseBody = errors.New("Response body is empty")
)

// Client interface wrap client implementation to perform the actual request
type Client interface {
	MakeRequest(r Request) *Result
}

// DefaultClient the default client which implement Client interface
type DefaultClient struct {
	Client  *http.Client
	Timeout int
}

// MakeRequest perform request to the endpoint based on `Request` parameter
// A successful call returns `Result` with Response field and nil error
// Otherwise a `Result` with error field will be return
func (c *DefaultClient) MakeRequest(r Request) *Result {

	result := &Result{}
	req, err := http.NewRequest(r.Method, r.URL, r.Body)
	if err != nil {
		result.Error = err
		return result
	}
	ctx := context.Background()
	if r.Context != nil {
		ctx = r.Context
	}

	host := "svc"
	uri, err := url.Parse(r.URL)
	if err != nil {
		host = uri.Hostname()
	}
	tr := otel.Tracer("golib/api")
	ctx, span := tr.Start(ctx, r.Method+" "+r.URL, trace.WithAttributes(semconv.PeerServiceKey.String(host)))
	defer span.End()

	req = req.WithContext(ctx)
	timeout := c.Timeout
	if r.Timeout != 0 {
		timeout = r.Timeout
	}
	if timeout != 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		defer cancel()
		req = req.WithContext(ctx)
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}
	if r.OnRequestStart != nil {
		err = r.OnRequestStart(req)
		if err != nil {
			result.Error = err
			return result
		}
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		span.RecordError(err)
		result.Error = err
		if r.OnRequestFinished != nil {
			r.OnRequestFinished(result)
		}
		return result
	}
	defer resp.Body.Close()
	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, resp.Body)
	if err != nil {
		span.RecordError(err)
		result.Error = custerr.ErrChain{
			Cause:   err,
			Message: "Failed copying response body to buffer",
		}
		return result
	}
	result.Response.StatusCode = resp.StatusCode
	result.Response.Body = buff
	if r.OnRequestFinished != nil {
		err = r.OnRequestFinished(result)
		if err != nil {
			result.Error = err
			return result
		}
	}

	return result
}

type Result struct {
	Response Response
	Error    error
}

// Consume will consume the result response body to the passed object
// If result contains error it will return the error
// Else if response status is not success it will return `ErrorStatusNotOK`
func (r *Result) Consume(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}
	if r.Response.StatusCode < 200 || r.Response.StatusCode > 299 {
		log.WithFields(log.Fields{
			"statusCode": r.Response.StatusCode,
			"response":   r.Response,
		}).Errorln("Error when make request")
		return &ErrorStatusNotOK{
			Response: r.Response,
		}
	}
	return r.Response.Consume(&v)
}

type Response struct {
	StatusCode int
	Body       *bytes.Buffer
}

// Consume will consume the body to the passed object, the body assumed to be a json object
func (resp *Response) Consume(v interface{}) error {
	if resp.Body == nil {
		return ErrEmptyResponseBody
	}

	err := json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		err = custerr.ErrChain{
			Cause: err,
			Fields: map[string]interface{}{
				"body": string(bodyBytes),
			},
			Message: "Failed copying response body to interface",
		}
	}

	return err
}

type ErrorStatusNotOK struct {
	Response
}

func (e *ErrorStatusNotOK) Error() string {
	body := ""
	if e.Body != nil {
		body = e.Body.String()
	}
	return fmt.Sprintf("Response return status not OK, with status code %d, and body: %s", e.StatusCode, body)
}
