package matcher

import (
	"github.com/diki-haryadi/govega/custerr"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func HaveErrorMessage(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(c custerr.ErrChain) string { return c.Message }, m)
}

func HaveErrorType(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(c custerr.ErrChain) error { return c.Type }, m)
}
