package scribe_test

import (
	"testing"

	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormattedMap(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("String", func() {
		it("returns a formatted string representation of the map", func() {
			Expect(scribe.FormattedMap{
				"third":  3,
				"first":  1,
				"second": 2,
			}.String()).To(Equal("first  -> 1\nsecond -> 2\nthird  -> 3"))
		})
	})
}
