package cargo_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDependencyResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		config cargo.Config
	)

	it.Before(func() {
		config = cargo.Config{
			Metadata: cargo.ConfigMetadata{
				Dependencies: []cargo.ConfigMetadataDependency{
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-1"},
						Version: "1.2.2",
					},
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-1"},
						Version: "1.2.3",
					},
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-1"},
						Version: "2.3.4",
					},
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-2"},
						Version: "1.2.3",
					},
					{
						ID:      "some-ignored-dependency",
						Stacks:  []string{"some-ignored-stack"},
						Version: "1.3.0",
					},
					{
						ID:      "dependency-B",
						Stacks:  []string{"stack-1"},
						Version: "1.2.4",
					},
					{
						ID:      "dependency-B",
						Stacks:  []string{"*"},
						Version: "2.3.4",
					},
				},
			},
		}
	})

	it("finds the exact matching dependency", func() {
		dependency, err := cargo.ResolveDependency(config, "dependency-A", "1.2.2", "stack-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
			ID:      "dependency-A",
			Stacks:  []string{"stack-1"},
			Version: "1.2.2",
		}))
	})

	context("when the dependency has a wildcard stack", func() {
		it("finds the highest matching version across all stack ids", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-B", "*", "unrecognized-stack")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-B",
				Stacks:  []string{"*"},
				Version: "2.3.4",
			}))
		})
	})

	context("when the dependency version is empty", func() {
		it("picks the dependency with the highest semantic version number", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-A", "", "stack-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-A",
				Stacks:  []string{"stack-1"},
				Version: "2.3.4",
			}))
		})
	})

	context("when the entry version is 'default'", func() {
		it("picks the dependency with the highest semantic version number", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-A", "default", "stack-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-A",
				Stacks:  []string{"stack-1"},
				Version: "2.3.4",
			}))
		})
	})

	context("when there is a version with a major, minor, patch, and pessimistic operator (~>)", func() {
		it("picks the dependency >= version and < major.minor+1", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-A", "~> 1.2.0", "stack-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-A",
				Stacks:  []string{"stack-1"},
				Version: "1.2.3",
			}))
		})
	})

	context("when there is a version with a major, minor, and pessimistic operator (~>)", func() {
		it("picks the dependency >= version and < major+1", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-A", "~> 1.1", "stack-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-A",
				Stacks:  []string{"stack-1"},
				Version: "1.2.3",
			}))
		})
	})

	context("when there is a version with a major line only and pessimistic operator (~>)", func() {
		it("picks the dependency >= version.0.0 and < major+1.0.0", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-A", "~> 1", "stack-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-A",
				Stacks:  []string{"stack-1"},
				Version: "1.2.3",
			}))
		})
	})

	context("when there is a default version in the config", func() {
		it.Before(func() {
			config.Metadata.DefaultVersions = map[string]string{"dependency-A": "1.2.x"}
		})

		context("when the dependency version is empty", func() {
			it("picks the highest-versioned dependency that matches the default version", func() {
				dependency, err := cargo.ResolveDependency(config, "dependency-A", "", "stack-1")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
					ID:      "dependency-A",
					Stacks:  []string{"stack-1"},
					Version: "1.2.3",
				}))
			})
		})

		context("when the dependency version is 'default'", func() {
			it("picks the highest-versioned dependency that matches the default version", func() {
				dependency, err := cargo.ResolveDependency(config, "dependency-A", "default", "stack-1")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
					ID:      "dependency-A",
					Stacks:  []string{"stack-1"},
					Version: "1.2.3",
				}))
			})
		})

		context("when the dependency version is provided", func() {
			it("picks the dependency that best matches the provided version instead of the default", func() {
				dependency, err := cargo.ResolveDependency(config, "dependency-A", "2.*", "stack-1")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
					ID:      "dependency-A",
					Stacks:  []string{"stack-1"},
					Version: "2.3.4",
				}))
			})
		})
	})

	context("when both a wildcard stack constraint and a specific stack constraint exist for the same dependency version", func() {
		it.Before(func() {
			config.Metadata.Dependencies = []cargo.ConfigMetadataDependency{
				{
					ID:      "dependency-A",
					Stacks:  []string{"stack-1"},
					Version: "1.2.1",
				},
				{
					ID:      "dependency-A",
					Stacks:  []string{"*"},
					Version: "1.2.1",
				},
				{
					ID:      "dependency-A",
					Stacks:  []string{"stack-1", "*"},
					Version: "1.2.3",
				},
				{
					ID:      "dependency-A",
					Stacks:  []string{"stack-1"},
					Version: "1.2.3",
				},
			}
		})

		it("selects the more specific stack constraint", func() {
			dependency, err := cargo.ResolveDependency(config, "dependency-A", "*", "stack-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				ID:      "dependency-A",
				Stacks:  []string{"stack-1"},
				Version: "1.2.3",
			}))
		})
	})

	context("failure cases", func() {
		context("when the entry version constraint is not valid", func() {
			it("returns an error", func() {
				_, err := cargo.ResolveDependency(config, "dependency-A", "this-is-not-semver", "stack-1")
				Expect(err).To(MatchError(ContainSubstring("improper constraint")))
			})
		})

		context("when the dependency version is not valid", func() {
			it.Before(func() {
				config.Metadata.Dependencies = []cargo.ConfigMetadataDependency{
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-1"},
						Version: "not semver",
					},
				}
			})

			it("returns an error", func() {
				_, err := cargo.ResolveDependency(config, "dependency-A", "1.2.3", "stack-1")
				Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
			})
		})

		context("when multiple dependencies have a wildcard stack for the same version", func() {
			it.Before(func() {
				config.Metadata.Dependencies = []cargo.ConfigMetadataDependency{
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-1", "*"},
						Version: "1.2.3",
					},
					{
						ID:      "dependency-A",
						Stacks:  []string{"stack-2", "*"},
						Version: "1.2.3",
					},
				}
			})

			it("returns an error", func() {
				_, err := cargo.ResolveDependency(config, "dependency-A", "1.2.3", "stack-1")
				Expect(err).To(MatchError(ContainSubstring(`multiple dependencies support wildcard stack for version: "1.2.3"`)))
			})
		})

		context("when the entry version constraint cannot be satisfied", func() {
			it("returns an error with all the supported versions listed", func() {
				_, err := cargo.ResolveDependency(config, "dependency-A", "9.9.9", "stack-1")
				Expect(err).To(MatchError(ContainSubstring("failed to satisfy \"dependency-A\" dependency version constraint \"9.9.9\": no compatible versions on \"stack-1\" stack. Supported versions are: [1.2.2, 1.2.3, 2.3.4]")))
			})
		})
	})
}
