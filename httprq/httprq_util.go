package httprq

import (
	"fmt"
	"strings"
	"time"
)

func BackOffDelay(delay time.Duration) DelayFn {
	return func(attempt int) time.Duration {
		return delay * (1 << (attempt - 1))
	}
}
func RetryError(rc *RequestConfig) string {
	logWithNumber := make([]string, len(rc.errorLog))
	for i, errLog := range rc.errorLog {
		if errLog != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s",
				i+1,
				errLog.Error(),
			)
		}
	}

	return fmt.Sprintf("fail:\n%s", strings.Join(logWithNumber, "\n"))
}
