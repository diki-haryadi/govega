package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	gock "gopkg.in/h2non/gock.v1"
)

const (
	endpoint = "http://my-endpoint.local"
)

func TestRequestBuilder_Execute(t *testing.T) {
	mClient := new(MockClient)
	mRetry := new(MockRetry)
	mRetry.On("DoRequest", mClient, mock.AnythingOfType("Request")).Return(Result{})
	Get("http://localhost.local").WithRetryStrategy(mRetry).WithClient(mClient).Execute()
	mRetry.AssertCalled(t, "DoRequest", mClient, mock.AnythingOfType("Request"))
}

func Test_Get_Success(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(endpoint).Get("/").Reply(204)
	result := Get(endpoint).Execute()
	g.Expect(result.Error).ShouldNot(HaveOccurred())
	g.Expect(result.Response.StatusCode).Should(Equal(204))
}

func Test_Post_Success(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(endpoint).Post("/").Reply(204)
	result := Post(endpoint).Execute()
	g.Expect(result.Error).ShouldNot(HaveOccurred())
	g.Expect(result.Response.StatusCode).Should(Equal(204))
}

func Test_Put_Success(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(endpoint).Put("/").Reply(204)
	result := Put(endpoint).Execute()
	g.Expect(result.Error).ShouldNot(HaveOccurred())
	g.Expect(result.Response.StatusCode).Should(Equal(204))
}

func Test_Delete_Success(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(endpoint).Delete("/").Reply(204)
	result := Delete(endpoint).Execute()
	g.Expect(result.Error).ShouldNot(HaveOccurred())
	g.Expect(result.Response.StatusCode).Should(Equal(204))
}

func TestRequestBuilder_WithBody(t *testing.T) {
	g := NewGomegaWithT(t)
	body := new(bytes.Buffer)
	rb := Get("").WithBody(body)
	g.Expect(rb.request.Body).Should(Equal(body))
}

func TestRequestBuilder_WithContext(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	rb := Get("").WithContext(ctx)
	g.Expect(rb.request.Context).Should(Equal(ctx))
}

func TestRequestBuilder_WithClient(t *testing.T) {
	g := NewGomegaWithT(t)
	cli := DefaultClient{}
	rb := Get("").WithClient(&cli)
	g.Expect(rb.client).Should(Equal(&cli))
}

func TestRequestBuilder_WithRetryStrategy(t *testing.T) {
	g := NewGomegaWithT(t)
	st := NoRetry{}
	rb := Get("").WithRetryStrategy(&st)
	g.Expect(rb.retryStrategy).Should(Equal(&st))
}

func TestRequestBuilder_WithTimeout(t *testing.T) {
	g := NewGomegaWithT(t)
	rb := Get("").WithTimeout(10)
	g.Expect(rb.request.Timeout).Should(Equal(10))
}

func TestRequestBuilder_AddHeader(t *testing.T) {
	g := NewGomegaWithT(t)
	rb := Get("").AddHeader("foo", "bar")
	g.Expect(rb.request.Headers).Should(HaveKeyWithValue("foo", "bar"))
}

func TestRequestBuilder_AddHeaders(t *testing.T) {
	g := NewGomegaWithT(t)
	rb := Get("").AddHeaders(map[string]string{
		"a": "b",
		"c": "d",
	})
	g.Expect(rb.request.Headers).Should(And(HaveKeyWithValue("a", "b"), HaveKeyWithValue("c", "d")))
}

func TestRequestBuilder_SetBasicAuth(t *testing.T) {
	username := "foo"
	password := "bar"
	auth := username + ":" + password
	g := NewGomegaWithT(t)
	rb := Get("").SetBasicAuth(username, password)
	g.Expect(rb.request.Headers).Should(HaveKeyWithValue("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth))))
}

func TestRequestBuilder_OnRequestStart(t *testing.T) {
	g := NewGomegaWithT(t)
	fn := func(*http.Request) error { return nil }
	rb := Get("").OnRequestStart(fn)
	g.Expect(rb.request.OnRequestStart).ShouldNot(BeNil())
}

func TestRequestBuilder_OnRequestFinished(t *testing.T) {
	g := NewGomegaWithT(t)
	fn := func(*Result) error { return nil }
	rb := Get("").OnRequestFinished(fn)
	g.Expect(rb.request.OnRequestFinished).ShouldNot(BeNil())
}
