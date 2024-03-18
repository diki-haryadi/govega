package event

import "errors"

var (
	ErrConsumerStarted = errors.New("[event/consumer] consumer already started")
)
