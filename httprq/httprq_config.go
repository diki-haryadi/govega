package httprq

import "time"

type DelayFn func(n int) time.Duration

type RequestConfig struct {
	NumRetry  int
	Delay     DelayFn
	errorLog  []error
	bodyBytes []byte
}
