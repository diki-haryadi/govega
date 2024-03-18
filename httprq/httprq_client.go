package httprq

import (
	"bytes"
	"context"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"go.elastic.co/apm/module/apmhttp"
	"io"
	"log"
	"net/http"
	"time"
	//"go.opentelemetry.io/otel"
	//semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	//"go.opentelemetry.io/otel/trace"
)

type RequestManager interface {
	MakeRequest(r Request) *Response
}

type RequestClient struct {
	Client  *http.Client
	Timeout int
}

func (rc *RequestClient) MakeRequest(r Request) *Response {

	res := &Response{}

	req, err := http.NewRequest(r.Method, r.URL, r.Body)
	if err != nil {
		res.Error = err
		return res
	}
	q := req.URL.Query()
	for k, v := range r.Params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	ctx := context.Background()
	if r.Context != nil {
		ctx = r.Context
	}

	req = req.WithContext(ctx)

	requestClientTimeout := rc.Timeout
	if r.Timeout != 0 {
		requestClientTimeout = r.Timeout
	}

	if requestClientTimeout != 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(requestClientTimeout))
		defer cancel()

		req = req.WithContext(ctx)
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	rc.Client = apmhttp.WrapClient(rc.Client)

	if r.TimeoutHystrix != 0 {
		hystrix.ConfigureCommand("http", hystrix.CommandConfig{
			Timeout:               r.TimeoutHystrix,
			MaxConcurrentRequests: r.MaxConcurrentRequests,
			ErrorPercentThreshold: r.ErrorPercentThreshold,
		})

		resultChan := make(chan *Response)
		errChan := hystrix.Go("http", func() error {
			result, err := rc.Client.Do(req)
			if err != nil {
				res.Error = err
				return err
			}

			buff := new(bytes.Buffer)
			_, err = io.Copy(buff, result.Body)
			if err != nil {
				res.Error = errors.New("failed copying response body to buffer")
				return err
			}

			if r.ErrNotSuccess {
				if res.StatusCode < 200 || res.StatusCode > 299 {
					res.Error = errors.New("error not success")
				}
			}

			res.StatusCode = result.StatusCode
			res.Body = buff
			defer result.Body.Close()
			resultChan <- res

			return err
		}, nil)

		select {
		case result := <-resultChan:
			log.Println("success:", result)
			return result

		case errors := <-errChan:
			log.Println("Circuit breaker opened for http command", errors.Error())
			return res
		}
	}

	responseClient, err := rc.Client.Do(req)
	if err != nil {
		res.Error = err
		return res
	}
	defer responseClient.Body.Close()

	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, responseClient.Body)
	if err != nil {

		res.Error = errors.New("failed copying response body to buffer")
		return res
	}

	if r.ErrNotSuccess {
		if res.StatusCode < 200 || res.StatusCode > 299 {
			res.Error = errors.New("error not success")
		}
	}

	res.StatusCode = responseClient.StatusCode
	res.Body = buff

	return res
}
