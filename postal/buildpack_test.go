package postal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestBuildpack(t *testing.T) {
	spec.Run(t, "Buildpack", testBuildpack, spec.Report(report.Terminal{}))
}

func testBuildpack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		path = filepath.Join(t.TempDir(), "buildpack.toml")
	})

	// parse writes the given buildpack.toml contents to a temporary file and
	// parses it back into a list of dependencies.
	parse := func(content string) []Dependency {
		Expect(os.WriteFile(path, []byte(content), 0600)).To(Succeed())

		dependencies, _, err := parseBuildpack(path, "")
		Expect(err).NotTo(HaveOccurred())

		return dependencies
	}

	context("standard format", func() {
		it("parses the standard dependency metadata keys", func() {
			eolDate, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

			dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  os = "linux"
  arch = "amd64"
  cpes = ["cpe:2.3:a:some:thing:1.2.3:*:*:*:*:*:*:*"]
  purls = ["pkg:generic/some-dependency@1.2.3"]
  eol-date = "2026-01-01T00:00:00Z"

  [[metadata.dependencies.distros]]
  name = "ubuntu"
  version = "22.04"

  [[metadata.dependencies.licenses]]
  type = "MIT"
  uri = "https://example.com/LICENSE"
`)

			Expect(dependencies).To(HaveLen(1))

			dependency := dependencies[0]
			Expect(dependency.ID).To(Equal("some-dependency"))
			Expect(dependency.CPEs).To(Equal([]string{"cpe:2.3:a:some:thing:1.2.3:*:*:*:*:*:*:*"}))
			Expect(dependency.PURLs).To(Equal([]string{"pkg:generic/some-dependency@1.2.3"}))
			Expect(dependency.OS).To(Equal("linux"))
			Expect(dependency.Arch).To(Equal("amd64"))
			Expect(dependency.EOLDate).To(Equal(eolDate))
			Expect(dependency.Distros).To(Equal([]Distro{{Name: "ubuntu", Version: "22.04"}}))
			Expect(dependency.LicenseEntries).To(Equal([]License{{Type: "MIT", URI: "https://example.com/LICENSE"}}))
		})

		it("allows the license uri to be omitted", func() {
			dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"

  [[metadata.dependencies.licenses]]
  type = "MIT"
`)

			Expect(dependencies).To(HaveLen(1))
			Expect(dependencies[0].LicenseEntries).To(Equal([]License{{Type: "MIT"}}))
		})

		it("populates both EOLDate and the deprecated DeprecationDate", func() {
			eolDate, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

			dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  eol-date = "2026-01-01T00:00:00Z"
`)

			Expect(dependencies).To(HaveLen(1))
			Expect(dependencies[0].EOLDate).To(Equal(eolDate))
			//nolint:staticcheck // SA1019: verifies the deprecated field stays in sync.
			Expect(dependencies[0].DeprecationDate).To(Equal(eolDate))
		})
	})

	// The backwards compatibility contexts below cover the dependency metadata
	// keys that predate the standard dependency metadata format (RFC 0059).
	// They exist to guarantee that buildpack.toml files written in the older
	// format keep working. Remove them once the legacy keys are no longer
	// supported.
	context("backwards compatibility", func() {
		context("when the dependency uses the legacy cpe key", func() {
			it("parses it into CPE", func() {
				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  cpe = "cpe:2.3:a:some:thing:1.2.3:*:*:*:*:*:*:*"
`)

				Expect(dependencies).To(HaveLen(1))
				//nolint:staticcheck // SA1019: verifies the deprecated field still parses.
				Expect(dependencies[0].CPE).To(Equal("cpe:2.3:a:some:thing:1.2.3:*:*:*:*:*:*:*"))
				Expect(dependencies[0].CPEs).To(BeEmpty())
			})
		})

		context("when the dependency uses the legacy purl key", func() {
			it("parses it into PURL", func() {
				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  purl = "pkg:generic/some-dependency@1.2.3"
`)

				Expect(dependencies).To(HaveLen(1))
				//nolint:staticcheck // SA1019: verifies the deprecated field still parses.
				Expect(dependencies[0].PURL).To(Equal("pkg:generic/some-dependency@1.2.3"))
				Expect(dependencies[0].PURLs).To(BeEmpty())
			})
		})

		context("when the dependency uses the legacy deprecation_date key", func() {
			it("parses it into DeprecationDate and EOLDate", func() {
				eolDate, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  deprecation_date = "2026-01-01T00:00:00Z"
`)

				Expect(dependencies).To(HaveLen(1))
				//nolint:staticcheck // SA1019: verifies the deprecated field still parses.
				Expect(dependencies[0].DeprecationDate).To(Equal(eolDate))
				Expect(dependencies[0].EOLDate).To(Equal(eolDate))
			})
		})

		context("when the dependency uses the legacy bare string licenses form", func() {
			it("parses them into Licenses", func() {
				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  licenses = ["MIT", "Apache-2.0"]
`)

				Expect(dependencies).To(HaveLen(1))
				Expect(dependencies[0].Licenses).To(Equal([]string{"MIT", "Apache-2.0"}))
				Expect(dependencies[0].LicenseEntries).To(BeEmpty())
			})
		})

		context("when the dependency uses the legacy OS key", func() {
			it("parses it into OS", func() {
				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  OS = "linux"
  arch = "amd64"
`)

				Expect(dependencies).To(HaveLen(1))
				Expect(dependencies[0].OS).To(Equal("linux"))
				Expect(dependencies[0].Arch).To(Equal("amd64"))
			})
		})

		context("when the dependency uses the legacy sha256, source_sha256, and stacks keys", func() {
			it("parses them", func() {
				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  sha256 = "some-sha"
  source_sha256 = "some-source-sha"
  source = "https://example.com/dependency.tar.gz.source"
  stacks = ["io.buildpacks.stacks.bionic"]
`)

				Expect(dependencies).To(HaveLen(1))
				Expect(dependencies[0].SHA256).To(Equal("some-sha"))
				Expect(dependencies[0].SourceSHA256).To(Equal("some-source-sha"))
				Expect(dependencies[0].Source).To(Equal("https://example.com/dependency.tar.gz.source"))
				Expect(dependencies[0].Stacks).To(Equal([]string{"io.buildpacks.stacks.bionic"}))
			})
		})

		context("when the dependency uses the dashed source-sha256 key", func() {
			it("parses it into SourceSHA256", func() {
				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  sha256 = "some-sha"
  source-sha256 = "some-source-sha"
`)

				Expect(dependencies).To(HaveLen(1))
				Expect(dependencies[0].SHA256).To(Equal("some-sha"))
				Expect(dependencies[0].SourceSHA256).To(Equal("some-source-sha"))
			})
		})

		context("when the dependency mixes the legacy and standard keys", func() {
			it("parses both", func() {
				eolDate, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				dependencies := parse(`
[[metadata.dependencies]]
  id = "some-dependency"
  version = "1.2.3"
  uri = "https://example.com/dependency.tar.gz"
  checksum = "sha256:some-sum"
  cpe = "cpe:2.3:a:some:thing:*:*:*:*:*:*:*:*"
  cpes = ["cpe:2.3:a:some:thing:1.2.3:*:*:*:*:*:*:*"]
  purl = "pkg:generic/some-dependency"
  purls = ["pkg:generic/some-dependency@1.2.3"]
  deprecation_date = "2026-01-01T00:00:00Z"
  eol-date = "2026-01-01T00:00:00Z"
`)

				Expect(dependencies).To(HaveLen(1))

				dependency := dependencies[0]
				//nolint:staticcheck // SA1019: verifies the deprecated field still parses.
				Expect(dependency.CPE).To(Equal("cpe:2.3:a:some:thing:*:*:*:*:*:*:*:*"))
				Expect(dependency.CPEs).To(Equal([]string{"cpe:2.3:a:some:thing:1.2.3:*:*:*:*:*:*:*"}))
				//nolint:staticcheck // SA1019: verifies the deprecated field still parses.
				Expect(dependency.PURL).To(Equal("pkg:generic/some-dependency"))
				Expect(dependency.PURLs).To(Equal([]string{"pkg:generic/some-dependency@1.2.3"}))
				Expect(dependency.EOLDate).To(Equal(eolDate))
			})
		})
	})
}
