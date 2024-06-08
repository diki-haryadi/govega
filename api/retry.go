package api

import (
	"fmt"
	"net"

	"bytes"
	"github.com/diki-haryadi/govega/custerr"
	"io/ioutil"
	"strings"
	"time"
)

type RetryStrategy interface {
	DoRequest(c Client, r Request) *Result
}

type DelayFunc func(n uint) time.Duration

type Config struct {
	NumRetry  uint
	DelayType DelayFunc
	errorLog  []error
	bodyBytes []byte
}

// NoRetry implement retry strategy which will no perform retry if request is not succeed
type NoRetry struct{}

func (nr *NoRetry) DoRequest(c Client, r Request) *Result {
	return c.MakeRequest(r)
}

// RetryIfTimeout implement retry strategy which will perform retry n time(s) if request timeout occurred
type RetryIfTimeout struct {
	NoRetry
	Config
}

func (rt *RetryIfTimeout) DoRequest(c Client, r Request) *Result {
	resp := rt.makeRequest(rt.Config.NumRetry, c, r)
	if resp.Error != nil {
		resp.Error = custerr.ErrChain{
			Cause:   resp.Error,
			Message: fmt.Sprintf("Timeout performing request to %s after %d attempts", r.URL, rt.NumRetry+1),
		}
	}
	return resp
}

func (rt *RetryIfTimeout) makeRequest(attempt uint, c Client, r Request) *Result {
	resp := rt.NoRetry.DoRequest(c, r)
	if resp.Error != nil {
		if errTimeout, ok := resp.Error.(net.Error); ok && errTimeout.Timeout() {
			if attempt > 0 {
				attempt--
				return rt.makeRequest(attempt, c, r)
			}
		}
	}
	return resp
}

type RetryAllErrors struct {
	NoRetry
	Config
}

func NewRetryAllErrors() *RetryAllErrors {
	rt := new(RetryAllErrors)
	rt.NumRetry = 3
	rt.DelayType = BackOffDelay(100 * time.Millisecond)
	return rt
}

func (rt *RetryAllErrors) DoRequest(c Client, r Request) *Result {

	resp := rt.makeRequest(rt.Config.NumRetry, c, r)
	if resp.Error != nil {
		resp.Error = custerr.ErrChain{
			Cause:   resp.Error,
			Message: fmt.Sprintf("Error performing request to %s after %d attempts %v", r.URL, rt.NumRetry+1, RetryError(&rt.Config)),
		}
	}
	return resp
}

func (rt *RetryAllErrors) makeRequest(attempt uint, c Client, r Request) *Result {
	if r.Body != nil {
		rt.bodyBytes, _ = ioutil.ReadAll(r.Body)
		r.Body = bytes.NewBuffer(rt.bodyBytes)
	}
	resp := rt.NoRetry.DoRequest(c, r)
	if resp.Error != nil {
		if attempt > 0 {
			rt.errorLog = append(rt.errorLog, resp.Error)
			time.Sleep(rt.DelayType(attempt))
			attempt--
			if rt.bodyBytes != nil {
				r.Body = ioutil.NopCloser(bytes.NewBuffer(rt.bodyBytes))
			}
			return rt.makeRequest(attempt, c, r)
		}
	}
	return resp
}

func BackOffDelay(delay time.Duration) DelayFunc {
	return func(attempt uint) time.Duration {
		return delay * (1 << (attempt - 1))
	}
}
func RetryError(config *Config) string {
	logWithNumber := make([]string, len(config.errorLog))
	for i, l := range config.errorLog {
		if l != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s", i+1, l.Error())
		}
	}

	return fmt.Sprintf("fail:\n%s", strings.Join(logWithNumber, "\n"))
}
