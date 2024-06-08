package api

import (
	"fmt"
	"testing"

	"github.com/diki-haryadi/govega/custerr"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	if cli, ok := defaultClient.(*DefaultClient); ok {
		gock.InterceptClient(cli.Client)
		defer gock.RestoreClient(cli.Client)
	}
}

type MockClient struct {
	mock.Mock
}

func (mc *MockClient) MakeRequest(r Request) *Result {
	args := mc.Called(r)
	var result Result
	var ok bool
	if result, ok = args.Get(0).(Result); !ok {
		panic(fmt.Sprintf("Invalid result arguments, expected type of Result but was %T", args.Get(0)))
	}
	return &result
}

type MockRetry struct {
	mock.Mock
}

func (mr *MockRetry) DoRequest(c Client, r Request) *Result {
	args := mr.Called(c, r)
	var result Result
	var ok bool
	if result, ok = args.Get(0).(Result); !ok {
		panic(fmt.Sprintf("Invalid result arguments, expected type of Result but was %T", args.Get(0)))
	}
	return &result
}

func HaveStatusCode(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(r Response) int { return r.StatusCode }, m)
}

func HaveBodyString(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(r Response) string { return string(r.Body.String()) }, m)
}

func HaveErrorMessage(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(c custerr.ErrChain) string { return c.Message }, m)
}

func HaveErrorType(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(c custerr.ErrChain) error { return c.Type }, m)
}
