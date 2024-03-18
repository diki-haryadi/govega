package httprq

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

type RequestRetry interface {
	DoRequest(rc RequestManager, r Request) *Response
}

type RequestNoRetry struct {
}

func (nr *RequestNoRetry) DoRequest(rc RequestManager, r Request) *Response {
	return rc.MakeRequest(r)
}

type RequestRetryWhenTimeout struct {
	RequestNoRetry
	RequestConfig
}

func (rt *RequestRetryWhenTimeout) DoRequest(rc RequestManager, r Request) *Response {
	res := rt.makeRequest(rc, r, rt.NumRetry)
	if res.Error != nil {
		res.Error = fmt.Errorf("timeout performing request to %s after %d attempts, error %s",
			r.URL,
			rt.NumRetry+1,
			res.Error.Error(),
		)
	}

	return res
}

func (rt *RequestRetryWhenTimeout) makeRequest(rc RequestManager, r Request, attempt int) *Response {
	res := rt.RequestNoRetry.DoRequest(rc, r)
	if res.Error != nil {
		if errTimeout, ok := res.Error.(net.Error); ok && errTimeout.Timeout() {
			if attempt > 0 {
				attempt--
				return rt.makeRequest(rc, r, attempt)
			}
		}
	}

	return res
}

type RequestRetryAllErrors struct {
	RequestNoRetry
	RequestConfig
}

func (rt *RequestRetryAllErrors) DoRequest(rc RequestManager, r Request) *Response {
	res := rt.makeRequest(rc, r, rt.NumRetry)
	if res.Error != nil {
		res.Error = fmt.Errorf("timeout performing request to %s after %d attempts, error %s, logNum %v",
			r.URL,
			rt.NumRetry+1,
			res.Error.Error(),
			RetryError(&rt.RequestConfig),
		)
	}

	return res
}

func (rt *RequestRetryAllErrors) makeRequest(rc RequestManager, r Request, attempt int) *Response {

	if r.Body != nil {
		rt.bodyBytes, _ = ioutil.ReadAll(r.Body)
		r.Body = bytes.NewBuffer(rt.bodyBytes)
	}

	res := rt.RequestNoRetry.DoRequest(rc, r)
	if res.Error != nil {
		if attempt > 0 {
			rt.errorLog = append(rt.errorLog, res.Error)
			time.Sleep(rt.Delay(attempt))

			attempt--
			if rt.bodyBytes != nil {
				r.Body = ioutil.NopCloser(bytes.NewBuffer(rt.bodyBytes))
			}

			return rt.makeRequest(rc, r, attempt)
		}
	}

	return res
}
