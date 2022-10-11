package postal_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/sclevine/spec"
)

func testBuildpack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "buildpack.toml")
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()
		_, err = file.WriteString(`
[[metadata.dependencies]]
deprecation_date = 2022-04-01T00:00:00Z
cpe = "some-cpe"
cpes = ["some-cpe", "other-cpe"]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-other-entry"
cpes = ["some-cpe", "other-cpe"]
sha256 = "some-other-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.4"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["other-stack"]
uri = "some-uri"
version = "1.2.5"

[[metadata.dependencies]]
id = "some-random-entry"
cpe = "some-cpe"
cpes = ["some-cpe", "other-cpe"]
sha256 = "some-random-sha"
stacks = ["other-random-stack"]
uri = "some-uri"
version = "1.3.0"

[[metadata.dependencies]]
id = "some-random-other-entry"
sha256 = "some-random-other-sha"
stacks = ["some-other-random-stack"]
uri = "some-uri"
version = "2.0.0"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "4.5.6"
strip-components = 1

[[metadata.dependencies]]
id = "some-other-entry"
sha256 = "some-sha"
stacks = ["*"]
uri = "some-uri"
version = "4.5.6"
strip-components = 1
`)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).To(Succeed())
	})

	it("finds the best matching dependency given a plan entry", func() {
		deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
		Expect(err).NotTo(HaveOccurred())

		dependency, err := postal.ResolveDependency(path, "some-entry", "1.2.*", "some-stack")
		Expect(err).NotTo(HaveOccurred())
		Expect(dependency).To(Equal(postal.Dependency{
			CPE:             "some-cpe",
			CPEs:            []string{"some-cpe", "other-cpe"},
			DeprecationDate: deprecationDate,
			ID:              "some-entry",
			Stacks:          []string{"some-stack"},
			URI:             "some-uri",
			SHA256:          "some-sha",
			Version:         "1.2.3",
		}))
	})

	context("when the dependency has a wildcard stack", func() {
		it("is compatible with all stack ids", func() {
			dependency, err := postal.ResolveDependency(path, "some-other-entry", "", "random-stack")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(postal.Dependency{
				ID:              "some-other-entry",
				Stacks:          []string{"*"},
				URI:             "some-uri",
				SHA256:          "some-sha",
				Version:         "4.5.6",
				StripComponents: 1,
			}))
		})
	})

	context("when there is NOT a default version", func() {
		context("when the entry version is empty", func() {
			it("picks the dependency with the highest semantic version number", func() {
				dependency, err := postal.ResolveDependency(path, "some-entry", "", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					ID:              "some-entry",
					Stacks:          []string{"some-stack"},
					URI:             "some-uri",
					SHA256:          "some-sha",
					Version:         "4.5.6",
					StripComponents: 1,
				}))
			})
		})

		context("when the entry version is default", func() {
			it("picks the dependency with the highest semantic version number", func() {
				dependency, err := postal.ResolveDependency(path, "some-entry", "default", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					ID:              "some-entry",
					Stacks:          []string{"some-stack"},
					URI:             "some-uri",
					SHA256:          "some-sha",
					Version:         "4.5.6",
					StripComponents: 1,
				}))
			})
		})

		context("when there is a version with a major, minor, patch, and pessimistic operator (~>)", func() {
			it("picks the dependency >= version and < major.minor+1", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				dependency, err := postal.ResolveDependency(path, "some-entry", "~> 1.2.0", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					DeprecationDate: deprecationDate,
					CPE:             "some-cpe",
					CPEs:            []string{"some-cpe", "other-cpe"},
					ID:              "some-entry",
					Stacks:          []string{"some-stack"},
					URI:             "some-uri",
					SHA256:          "some-sha",
					Version:         "1.2.3",
				}))
			})
		})

		context("when there is a version with a major, minor, and pessimistic operator (~>)", func() {
			it("picks the dependency >= version and < major+1", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				dependency, err := postal.ResolveDependency(path, "some-entry", "~> 1.1", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					CPE:             "some-cpe",
					CPEs:            []string{"some-cpe", "other-cpe"},
					DeprecationDate: deprecationDate,
					ID:              "some-entry",
					Stacks:          []string{"some-stack"},
					URI:             "some-uri",
					SHA256:          "some-sha",
					Version:         "1.2.3",
				}))
			})
		})

		context("when there is a version with a major line only and pessimistic operator (~>)", func() {
			it("picks the dependency >= version.0.0 and < major+1.0.0", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				dependency, err := postal.ResolveDependency(path, "some-entry", "~> 1", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					CPE:             "some-cpe",
					CPEs:            []string{"some-cpe", "other-cpe"},
					DeprecationDate: deprecationDate,
					ID:              "some-entry",
					Stacks:          []string{"some-stack"},
					URI:             "some-uri",
					SHA256:          "some-sha",
					Version:         "1.2.3",
				}))
			})
		})
	})

	context("when there is a default version", func() {
		it.Before(func() {
			err := os.WriteFile(path, []byte(`
[metadata]
[metadata.default-versions]
some-entry = "1.2.x"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-other-entry"
sha256 = "some-other-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.4"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["other-stack"]
uri = "some-uri"
version = "1.2.5"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "4.5.6"
`), 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		context("when the entry version is empty", func() {
			it("picks the dependency that best matches the default version", func() {
				dependency, err := postal.ResolveDependency(path, "some-entry", "", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					ID:      "some-entry",
					Stacks:  []string{"some-stack"},
					URI:     "some-uri",
					SHA256:  "some-sha",
					Version: "1.2.3",
				}))
			})
		})

		context("when the entry version is default", func() {
			it("picks the dependency that best matches the default version", func() {
				dependency, err := postal.ResolveDependency(path, "some-entry", "default", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					ID:      "some-entry",
					Stacks:  []string{"some-stack"},
					URI:     "some-uri",
					SHA256:  "some-sha",
					Version: "1.2.3",
				}))
			})
		})
	})

	context("when both a wildcard stack constraint and a specific stack constraint exist for the same dependency version", func() {
		it.Before(func() {
			err := os.WriteFile(path, []byte(`
[metadata]
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri-specific-stack"
version = "1.2.1"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["*"]
uri = "some-uri-only-wildcard"
version = "1.2.1"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack","*"]
uri = "some-uri-only-wildcard"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri-specific-stack"
version = "1.2.3"
`), 0600)

			Expect(err).NotTo(HaveOccurred())
		})

		it("selects the more specific stack constraint", func() {
			dependency, err := postal.ResolveDependency(path, "some-entry", "*", "some-stack")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(postal.Dependency{
				ID:      "some-entry",
				Stacks:  []string{"some-stack"},
				URI:     "some-uri-specific-stack",
				SHA256:  "some-sha",
				Version: "1.2.3",
			}))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.toml is malformed", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte("this is not toml"), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := postal.ResolveDependency(path, "some-entry", "1.2.3", "some-stack")
				Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack.toml")))
			})
		})

		context("when the entry version constraint is not valid", func() {
			it("returns an error", func() {
				_, err := postal.ResolveDependency(path, "some-entry", "this-is-not-semver", "some-stack")
				Expect(err).To(MatchError(ContainSubstring("improper constraint")))
			})
		})

		context("when the dependency version is not valid", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte(`
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "this is super not semver"
`), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := postal.ResolveDependency(path, "some-entry", "1.2.3", "some-stack")
				Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
			})
		})

		context("when multiple dependencies have a wildcard stack for the same version", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte(`
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha-A"
stacks = ["some-stack","*"]
uri = "some-uri-A"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha-B"
stacks = ["some-stack","some-other-stack","*"]
uri = "some-uri-B"
version = "1.2.3"
`), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := postal.ResolveDependency(path, "some-entry", "1.2.3", "some-stack")
				Expect(err).To(MatchError(ContainSubstring(`multiple dependencies support wildcard stack for version: "1.2.3"`)))
			})
		})

		context("when the entry version constraint cannot be satisfied", func() {
			it("returns an error with all the supported versions listed", func() {
				_, err := postal.ResolveDependency(path, "some-entry", "9.9.9", "some-stack")
				Expect(err).To(MatchError(ContainSubstring("failed to satisfy \"some-entry\" dependency version constraint \"9.9.9\": no compatible versions on \"some-stack\" stack. Supported versions are: [1.2.3, 4.5.6]")))
			})
		})
	})
}
