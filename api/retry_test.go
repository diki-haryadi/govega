package api

import (
	"net"
	"testing"

	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func TestNoRetry_DoRequest(t *testing.T) {
	mClient := new(MockClient)
	req := Request{}
	mClient.On("MakeRequest", req).Return(Result{})
	nr := NoRetry{}
	nr.DoRequest(mClient, req)
	mClient.AssertExpectations(t)
}

func TestRetryIfTimeout_DoRequest_Succeed(t *testing.T) {
	g := NewGomegaWithT(t)
	mClient := new(MockClient)
	req := Request{}
	mClient.On("MakeRequest", req).Return(Result{})
	rt := RetryIfTimeout{}
	result := rt.DoRequest(mClient, req)
	mClient.AssertExpectations(t)
	g.Expect(result.Error).ShouldNot(HaveOccurred())
}

func TestRetryIfTimeout_DoRequest_KeepRetryingWhenTimeoutOccurred(t *testing.T) {
	g := NewGomegaWithT(t)
	mClient := new(MockClient)
	req := Request{URL: testEndpoint}
	err := &net.DNSError{
		IsTimeout: true,
	}
	mClient.On("MakeRequest", req).Return(Result{Error: err})
	rt := RetryIfTimeout{}
	rt.NumRetry = 3
	result := rt.DoRequest(mClient, req)
	mClient.AssertNumberOfCalls(t, "MakeRequest", 4)
	g.Expect(result.Error).Should(SatisfyAll(HaveOccurred(),
		HaveErrorMessage(Equal("Timeout performing request to http://www.sicepat.com after 4 attempts"))))
}

func TestRetryIfTimeout_DoRequest_NotRetryingWhenErrorIsNotTimeout(t *testing.T) {
	g := NewGomegaWithT(t)
	mClient := new(MockClient)
	req := Request{}
	err := errors.New("Test Error")
	mClient.On("MakeRequest", req).Return(Result{Error: err})
	rt := RetryIfTimeout{}
	rt.NumRetry = 3
	result := rt.DoRequest(mClient, req)
	mClient.AssertNumberOfCalls(t, "MakeRequest", 1)
	g.Expect(result.Error).Should(HaveOccurred())
}

func TestRetryAllErrors_DoRequest_Succeed(t *testing.T) {
	g := NewGomegaWithT(t)
	mClient := new(MockClient)
	req := Request{URL: testEndpoint}
	mClient.On("MakeRequest", req).Return(Result{})
	rt := RetryAllErrors{}
	result := rt.DoRequest(mClient, req)
	mClient.AssertExpectations(t)
	g.Expect(result.Error).ShouldNot(HaveOccurred())
}

func TestRetryAllErrors_DoRequest_KeepRetryingWhenErrorOccurred(t *testing.T) {
	g := NewGomegaWithT(t)
	mClient := new(MockClient)
	req := Request{URL: testEndpoint}
	err := errors.New("something shit happens")
	mClient.On("MakeRequest", req).Return(Result{Error: err})
	rt := RetryAllErrors{}
	rt.NumRetry = 3
	rt.Config.DelayType = BackOffDelay(100 * time.Millisecond)
	result := rt.DoRequest(mClient, req)
	mClient.AssertNumberOfCalls(t, "MakeRequest", 4)
	g.Expect(result.Error).Should(SatisfyAll(HaveOccurred(),
		HaveErrorMessage(ContainSubstring("Error performing request to http://www.sicepat.com after 4 attempts fail"))))
}

func TestRetryAllErrorsWithDefaultValue_DoRequest_KeepRetryingWhenErrorOccurred(t *testing.T) {
	g := NewGomegaWithT(t)
	mClient := new(MockClient)
	req := Request{URL: testEndpoint}
	err := errors.New("something shit happens")
	mClient.On("MakeRequest", req).Return(Result{Error: err})
	rt := NewRetryAllErrors()
	result := rt.DoRequest(mClient, req)
	mClient.AssertNumberOfCalls(t, "MakeRequest", 4)
	g.Expect(result.Error).Should(SatisfyAll(HaveOccurred(),
		HaveErrorMessage(ContainSubstring("Error performing request to http://www.sicepat.com after 4 attempts fail"))))
}
