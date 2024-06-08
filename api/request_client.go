package api

import (
	"fmt"
	"net/http"
)

type APIRequestClient struct {
	name                string
	baseURL             string
	timeout             int
	onRequestStartFn    OnRequestStartFunc
	onRequestFinishedFn OnRequestFinishedFunc
}

func NewAPIRequestClient(name string, baseURL string, timeout int) *APIRequestClient {
	return &APIRequestClient{
		name:    name,
		baseURL: baseURL,
		timeout: timeout,
	}
}

func (arc *APIRequestClient) Do(method, url string) *RequestBuilder {
	return Do(method, fmt.Sprintf("%s%s", arc.baseURL, url)).WithTimeout(arc.timeout)
}

func (arc *APIRequestClient) OnRequestStart(fn OnRequestStartFunc) *APIRequestClient {
	arc.onRequestStartFn = fn
	return arc
}

func (arc *APIRequestClient) OnRequestFinished(fn OnRequestFinishedFunc) *APIRequestClient {
	arc.onRequestFinishedFn = fn
	return arc
}

func (arc *APIRequestClient) Post(url string) *RequestBuilder {
	return arc.Do(http.MethodPost, url)
}

func (arc *APIRequestClient) Get(url string) *RequestBuilder {
	return arc.Do(http.MethodGet, url)
}

func (arc *APIRequestClient) Put(url string) *RequestBuilder {
	return arc.Do(http.MethodPut, url)
}

func (arc *APIRequestClient) Delete(url string) *RequestBuilder {
	return arc.Do(http.MethodDelete, url)
}
