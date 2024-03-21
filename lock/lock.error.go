package lock

import "errors"

var (
	ErrResourceLocked = errors.New("resource locked")
)
