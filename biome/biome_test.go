package biome_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/biome"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBiome(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("GetBool", func() {
		it("returns false when the environemnt variable isn't set", func() {
			ok, err := biome.GetBool("BOOL_VAR")
			Expect(err).NotTo(HaveOccurred())

			Expect(ok).To(BeFalse())
		})

		context("when the environment variable is set to something truthy", func() {
			it.Before(func() {
				Expect(os.Setenv("BOOL_VAR", "TRUE")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BOOL_VAR")).To(Succeed())
			})

			it("returns true", func() {
				ok, err := biome.GetBool("BOOL_VAR")
				Expect(err).NotTo(HaveOccurred())

				Expect(ok).To(BeTrue())
			})
		})

		context("when the environment variable is set to something falsey", func() {
			it.Before(func() {
				Expect(os.Setenv("BOOL_VAR", "FALSE")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BOOL_VAR")).To(Succeed())
			})

			it("returns false", func() {
				ok, err := biome.GetBool("BOOL_VAR")
				Expect(err).NotTo(HaveOccurred())

				Expect(ok).To(BeFalse())
			})
		})

		context("failure cases", func() {
			context("when the environment variable is set to something invalid", func() {
				it.Before(func() {
					Expect(os.Setenv("BOOL_VAR", "not-bool")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BOOL_VAR")).To(Succeed())
				})

				it("returns an error", func() {
					_, err := biome.GetBool("BOOL_VAR")
					Expect(err).To(MatchError("invalid value 'not-bool' for key 'BOOL_VAR': expected one of [1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False]"))
				})
			})
		})
	})
}
