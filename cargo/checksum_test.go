package cargo_test

import (
	"fmt"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testChecksum(t *testing.T, context spec.G, it spec.S) {
	Expect := NewWithT(t).Expect

	context("Matching", func() {
		type testCaseType struct {
			c1     string
			c2     string
			result bool
		}

		for _, tc := range []testCaseType{
			{"", "", true},
			{"c", "c", true},
			{"sha256:c", "c", true},
			{"c", "sha256:c", true},
			{"md5:c", "md5:c", true},
			{"md5:c", ":c", false},
			{":", ":", true},
			{":c", ":c", true},
			{"", "c", false},
			{"c", "", false},
			{"c", "z", false},
			{"md5:c", "sha256:c", false},
			{"md5:c", "md5:d", false},
			{"md5:c:d", "md5:c:d", true},
			{"md5:c", "md5:c:d", false},
			{":", "::", false},
			{":", ":::", false},
		} {

			// NOTE: we need to keep a "loop-local" variable to use in the "it
			// function closure" below, otherwise the value of tc will simply be the
			// last element in the slice every time the test is evaluated.
			ca, cb, sb, result := cargo.Checksum(tc.c1), cargo.Checksum(tc.c2), tc.c2, tc.result

			it(fmt.Sprintf("will check result %q == %q", ca, cb), func() {
				Expect(ca.Match(cb)).To(Equal(result))
				Expect(ca.MatchString(sb)).To(Equal(result))
			})
		}
	})
}
