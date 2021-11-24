package packit_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEmptySBOM(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Format", func() {
		var entries packit.SBOMEntries

		it.Before(func() {
			entries = packit.EmptySBOM{}
		})

		it("Returns an empty map", func() {
			result := entries.Format()

			Expect(result).NotTo(BeNil())
			Expect(result).To(BeEmpty())
		})
	})
	context("IsEmpty", func() {
		var entries packit.SBOMEntries

		it.Before(func() {
			entries = packit.EmptySBOM{}
		})

		it("Returns true", func() {
			result := entries.IsEmpty()

			Expect(result).To(BeTrue())
		})
	})
}
