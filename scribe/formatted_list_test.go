package scribe_test

import (
	"testing"

	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormattedList(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("String", func() {
		it("returns a formatted string representation of the list", func() {
			Expect(scribe.FormattedList{
				"third",
				"first",
				"second",
			}.String()).To(Equal("first\nsecond\nthird"))
		})
	})
}
