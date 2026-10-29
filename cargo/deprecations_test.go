package cargo_test

import (
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDeprecations(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("DeprecatedDependencyKeys", func() {
		it("returns nothing when there are no dependencies", func() {
			deprecations, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`
[buildpack]
	id = "some-buildpack-id"
`))
			Expect(err).NotTo(HaveOccurred())
			Expect(deprecations).To(BeEmpty())
		})

		it("returns nothing when the dependencies use the standard format", func() {
			deprecations, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	version = "1.2.3"
	uri = "https://example.com/dependency.tar.gz"
	checksum = "sha256:some-sum"
	source-checksum = "sha256:some-source-sum"
	os = "linux"
	arch = "amd64"
	cpes = ["some-cpe"]
	purls = ["some-purl"]
	eol-date = "2020-06-01T00:00:00Z"

	[[metadata.dependencies.distros]]
	name = "ubuntu"
	version = "22.04"

	[[metadata.dependencies.licenses]]
	type = "MIT"
	uri = "https://example.com/LICENSE"
`))
			Expect(err).NotTo(HaveOccurred())
			Expect(deprecations).To(BeEmpty())
		})

		it("reports each deprecated dependency key", func() {
			deprecations, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	version = "1.2.3"
	cpe = "some-cpe"
	purl = "some-purl"
	deprecation_date = "2020-06-01T00:00:00Z"
	sha256 = "some-sha"
	source_sha256 = "some-source-sha"
	stacks = ["some-stack"]
	licenses = ["MIT"]
`))
			Expect(err).NotTo(HaveOccurred())

			Expect(deprecations).To(Equal([]cargo.Deprecation{
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "cpe", Replacement: "cpes"},
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "deprecation_date", Replacement: "eol-date"},
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "licenses", Replacement: "licenses tables with type and uri"},
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "purl", Replacement: "purls"},
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "sha256", Replacement: "checksum"},
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "source_sha256", Replacement: "source-checksum"},
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "stacks", Replacement: "distros, os, and arch"},
			}))
		})

		it("reports the dashed source-sha256 spelling", func() {
			deprecations, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	source-sha256 = "some-source-sha"
`))
			Expect(err).NotTo(HaveOccurred())
			Expect(deprecations).To(Equal([]cargo.Deprecation{
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "source-sha256", Replacement: "source-checksum"},
			}))
		})

		it("reports deprecations per dependency in document order", func() {
			deprecations, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`
[[metadata.dependencies]]
	id = "first"
	cpe = "some-cpe"

[[metadata.dependencies]]
	id = "second"
	version = "1.2.3"

[[metadata.dependencies]]
	id = "third"
	purl = "some-purl"
`))
			Expect(err).NotTo(HaveOccurred())
			Expect(deprecations).To(Equal([]cargo.Deprecation{
				{Path: "metadata.dependencies[0]", ID: "first", Key: "cpe", Replacement: "cpes"},
				{Path: "metadata.dependencies[2]", ID: "third", Key: "purl", Replacement: "purls"},
			}))
		})

		it("works for extension.toml documents", func() {
			deprecations, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`
api = "0.7"

[extension]
	id = "some-extension-id"

[[metadata.dependencies]]
	id = "some-dependency"
	cpe = "some-cpe"
`))
			Expect(err).NotTo(HaveOccurred())
			Expect(deprecations).To(Equal([]cargo.Deprecation{
				{Path: "metadata.dependencies[0]", ID: "some-dependency", Key: "cpe", Replacement: "cpes"},
			}))
		})

		context("failure cases", func() {
			context("when the document is not valid TOML", func() {
				it("returns an error", func() {
					_, err := cargo.DeprecatedDependencyKeys(strings.NewReader(`not = valid = toml`))
					Expect(err).To(MatchError(ContainSubstring("expected value but found")))
				})
			})
		})
	})
}
