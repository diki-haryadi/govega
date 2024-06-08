package taskworker

import (
	"errors"
	"sync"
	"sync/atomic"
)

type holder struct {
	mx     sync.RWMutex
	active int32
	res    []Result
}

func (h *holder) Add() {
	atomic.AddInt32(&h.active, 1)
}

func (h *holder) GetActiveWorker() int32 {
	return atomic.LoadInt32(&h.active)
}

func (h *holder) Store(result interface{}, err error) {
	h.mx.Lock()
	defer h.mx.Unlock()

	h.res = append(h.res, Result{Result: result, Err: err})
	atomic.AddInt32(&h.active, -1)
}

func (h *holder) GetAllResult() []Result {
	h.mx.RLock()
	defer h.mx.RUnlock()

	return h.res
}

type Result struct {
	Result interface{}
	Err    error
}

var ErrorInvalidObject = errors.New("InvalidObject")
