package paketosbom_test

import (
	"testing"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	//nolint Ignore SA1019, usage of deprecated package within a deprecated test case
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
)

func testPaketoSBOM(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("GetBOMChecksumAlgorithm", func() {
		context("given an algorithm string with the exact name of a CycloneDX algorithm", func() {
			it("returns the same algorithm name", func() {
				algorithm512, err := paketosbom.GetBOMChecksumAlgorithm("SHA-512")
				Expect(err).ToNot(HaveOccurred())
				Expect(algorithm512).To(Equal(paketosbom.SHA512))
			})
		})
		context("given an algorithm string with a lowercase version of a CycloneDX algorithm", func() {
			it("returns the Cyclonedx algorithm name", func() {
				algorithm512, err := paketosbom.GetBOMChecksumAlgorithm("sha-512")
				Expect(err).ToNot(HaveOccurred())
				Expect(algorithm512).To(Equal(paketosbom.SHA512))
			})
			context("it also does not contain a dash", func() {
				it("returns the Cyclonedx algorithm name", func() {
					algorithm512, err := paketosbom.GetBOMChecksumAlgorithm("sha512")
					Expect(err).ToNot(HaveOccurred())
					Expect(algorithm512).To(Equal(paketosbom.SHA512))
				})
			})
		})
		context("failure cases", func() {
			context("when the attempted BOM checksum algorithm is not supported", func() {
				it("persists a build.toml", func() {
					alg, err := paketosbom.GetBOMChecksumAlgorithm("RANDOM-ALG")
					Expect(alg).To(Equal(paketosbom.UNKNOWN))
					Expect(err).To(MatchError("failed to get supported BOM checksum algorithm: RANDOM-ALG is not valid"))
				})
			})
		})
	})
}
