package cargo_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"
)

func testAreChecksumsEqual(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("AreChecksumsEqual", func() {
		type testCaseType struct {
			c1       string
			c2       string
			expected bool
		}

		for _, testCase := range []testCaseType{
			{"", "", true},
			{"c", "c", true},
			{"c", "c", true},
			{"sha256:c", "c", true},
			{"c", "sha256:c", true},
			{"md5:c", "md5:c", true},
			{":", ":", true},
			{":c", ":c", true},
			{"", "c", false},
			{"c", "", false},
			{"c", "z", false},
			{"md5:c", "sha256:c", false},
			{"md5:c:d", "md5:c:d", false},
			{"md5:c", "md5:c:d", false},
			{":", "::", false},
			{":", ":::", false},
		} {
			it("will check result", func() {
				Expect(cargo.AreChecksumsEqual(testCase.c1, testCase.c2)).To(Equal(testCase.expected))
			})
		}
	})
}
