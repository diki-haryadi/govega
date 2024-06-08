package mock

import (
	"fmt"

	"github.com/diki-haryadi/govega/api"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (mc *MockClient) MakeRequest(r api.Request) *api.Result {
	args := mc.Called(r)
	var result api.Result
	var ok bool
	if result, ok = args.Get(0).(api.Result); !ok {
		panic(fmt.Sprintf("Invalid result arguments, expected type of api.Result but was %T", args.Get(0)))
	}
	return &result
}

type MockRetry struct {
	mock.Mock
}

func (mr *MockRetry) DoRequest(c api.Client, r api.Request) *api.Result {
	args := mr.Called(c, r)
	var result api.Result
	var ok bool
	if result, ok = args.Get(0).(api.Result); !ok {
		panic(fmt.Sprintf("Invalid result arguments, expected type of api.Result but was %T", args.Get(0)))
	}
	return &result
}
