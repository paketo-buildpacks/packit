package internal_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit/v2/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFail(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("acts as an error", func() {
		fail := internal.Fail
		Expect(fail).To(MatchError("failed"))
	})

	context("when given a message", func() {
		it("acts as an error with that message", func() {
			fail := internal.Fail.WithMessage("this is a %s", "failure message")
			Expect(fail).To(MatchError("this is a failure message"))
		})
	})
}
