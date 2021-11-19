package packit_test

import (
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSBOMEntries(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Set", func() {
		var entries packit.SBOMEntries

		it.Before(func() {
			entries = packit.SBOMEntries{}
		})

		it("initializes an empty layer", func() {
			entries.Set("some-name.some-extension", strings.NewReader("some-content"))
			entries.Set("other-name.other-extension", strings.NewReader("other-content"))

			Expect(entries).To(Equal(packit.SBOMEntries{
				"some-name.some-extension":   strings.NewReader("some-content"),
				"other-name.other-extension": strings.NewReader("other-content"),
			}))
		})
	})
}
