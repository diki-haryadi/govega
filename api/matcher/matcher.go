package matcher

import (
	"github.com/diki-haryadi/govega/api"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func HaveStatusCode(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(r api.Response) int { return r.StatusCode }, m)
}

func HaveBodyString(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(r api.Response) string { return string(r.Body.Bytes()) }, m)
}
