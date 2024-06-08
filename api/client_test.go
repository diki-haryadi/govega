package api

import (
	"bytes"
	"net"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

const (
	testEndpoint = "http://www.sicepat.com"
)

func TestDefaultClient_MakeRequest_ContainsResponseWhenSucceed(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(testEndpoint).Get("/").Reply(200).BodyString("Hello World!")
	c := newTestClient()
	result := c.MakeRequest(Request{
		URL: testEndpoint,
	})
	g.Expect(result.Error).ShouldNot(HaveOccurred())
	g.Expect(result.Response).Should(SatisfyAll(HaveStatusCode(Equal(200)),
		HaveBodyString(Equal("Hello World!"))))
}

func TestDefaultClient_MakeRequest_SetHeader(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(testEndpoint).Get("/").MatchHeader("foo", "bar").Reply(200)
	c := newTestClient()
	result := c.MakeRequest(Request{
		URL: testEndpoint,
		Headers: map[string]string{
			"foo": "bar",
		},
	})
	g.Expect(result.Error).ShouldNot(HaveOccurred())
}

func TestDefaultClient_MakeRequest_ReturnTimeoutErrorWhenExceedingDuration(t *testing.T) {
	g := NewGomegaWithT(t)
	c := newTestClient()
	result := c.MakeRequest(Request{
		URL:     "http://slowwly.robertomurray.co.uk/delay/3000/url/http://www.google.com",
		Timeout: 1,
	})
	g.Expect(result.Error).Should(SatisfyAll(HaveOccurred(),
		WithTransform(func(err net.Error) bool { return err.Timeout() }, BeTrue())))
}

func TestDefaultClient_MakeRequest_InvalidMethod(t *testing.T) {
	g := NewGomegaWithT(t)
	c := newTestClient()
	result := c.MakeRequest(Request{
		Method:  "bad method",
		URL:     "https://www.google.com",
		Timeout: 1,
	})
	g.Expect(result.Error).Should(HaveOccurred())
}

func TestDefaultClient_MakeRequest_OnRequestStart(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(testEndpoint).Get("/").Reply(200).BodyString("Hello World!")
	c := newTestClient()
	var isCalled bool
	c.MakeRequest(Request{
		URL: testEndpoint,
		OnRequestStart: func(_ *http.Request) error {
			g.Expect(gock.IsPending()).Should(BeTrue())
			isCalled = true
			return nil
		},
	})
	g.Expect(isCalled).Should(BeTrue())
	g.Expect(gock.IsPending()).Should(BeFalse())
}

func TestDefaultClient_MakeRequest_OnRequestFinished(t *testing.T) {
	g := NewGomegaWithT(t)
	defer gock.Off()
	gock.New(testEndpoint).Get("/").Reply(200).BodyString("Hello World!")
	c := newTestClient()
	var isCalled bool
	c.MakeRequest(Request{
		URL: testEndpoint,
		OnRequestFinished: func(r *Result) error {
			g.Expect(gock.IsDone()).Should(BeTrue())
			isCalled = true
			return nil
		},
	})
	g.Expect(isCalled).Should(BeTrue())
}

func TestResult_Consume_ReturnErrorWhenErrorPresent(t *testing.T) {
	g := NewGomegaWithT(t)
	result := Result{
		Error: errors.New("Test Error"),
	}
	g.Expect(result.Consume(map[string]interface{}{})).ShouldNot(Succeed())
}

func TestResponse_Consume_ReturnErrorWhenBodyIsEmpty(t *testing.T) {
	g := NewGomegaWithT(t)
	resp := Response{}
	g.Expect(resp.Consume(map[string]interface{}{})).Should(SatisfyAll(Not(Succeed()),
		Equal(ErrEmptyResponseBody)))
}

func TestResult_Consume_Succeed(t *testing.T) {
	g := NewGomegaWithT(t)
	result := Result{
		Response: Response{
			StatusCode: 200,
			Body:       bytes.NewBufferString(`{"foo":"bar"}`),
		},
	}
	m := map[string]interface{}{}
	g.Expect(result.Consume(&m)).Should(Succeed())
	g.Expect(m["foo"]).Should(Equal("bar"))
}

func TestResponse_Consume_Succeed(t *testing.T) {
	g := NewGomegaWithT(t)
	resp := Response{
		Body: bytes.NewBufferString(`{"foo":"bar"}`),
	}
	m := map[string]interface{}{}
	g.Expect(resp.Consume(&m)).Should(Succeed())
	g.Expect(m["foo"]).Should(Equal("bar"))
}

func TestErrorStatusNotOK_Error_EmptyBody(t *testing.T) {
	g := NewGomegaWithT(t)
	err := ErrorStatusNotOK{
		Response: Response{
			StatusCode: http.StatusBadRequest,
		},
	}
	g.Expect(err.Error()).Should(Equal("Response return status not OK, with status code 400, and body: "))
}

func TestErrorStatusNotOK_Error_WithBody(t *testing.T) {
	g := NewGomegaWithT(t)
	err := ErrorStatusNotOK{
		Response: Response{
			StatusCode: http.StatusBadRequest,
			Body:       bytes.NewBufferString("Hello World!"),
		},
	}
	g.Expect(err.Error()).Should(Equal("Response return status not OK, with status code 400, and body: Hello World!"))
}

func newTestClient() *DefaultClient {
	return &DefaultClient{
		Client: &http.Client{},
	}
}
