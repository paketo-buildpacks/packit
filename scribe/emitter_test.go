package scribe_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter scribe.Emitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = scribe.NewEmitter(buffer)
	})

	context("SelectedDependency", func() {
		it("prints details about the selected dependency", func() {
			entry := packit.BuildpackPlanEntry{
				Metadata: map[string]interface{}{"version-source": "some-source"},
			}
			dependency := postal.Dependency{
				Name:    "Some Dependency",
				Version: "some-version",
			}

			emitter.SelectedDependency(entry, dependency, time.Now())
			Expect(buffer.String()).To(Equal("    Selected Some Dependency version (using some-source): some-version\n\n"))
		})

		context("when the version source is missing", func() {
			it("prints details about the selected dependency", func() {
				dependency := postal.Dependency{
					Name:    "Some Dependency",
					Version: "some-version",
				}

				emitter.SelectedDependency(packit.BuildpackPlanEntry{}, dependency, time.Now())
				Expect(buffer.String()).To(Equal("    Selected Some Dependency version (using <unknown>): some-version\n\n"))
			})
		})

		context("when it is within 30 days of the deprecation date", func() {
			it("returns a warning that the dependency will be deprecated after the deprecation date", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate.Add(-29 * 24 * time.Hour)

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Some Dependency",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Some Dependency version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Some Dependency will be deprecated after 2021-04-01.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Some Dependency before this time.\n\n"))
			})
		})

		context("when it is on the the deprecation date", func() {
			it("returns a warning that the version of the dependency is no longer supported", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Some Dependency",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Some Dependency version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Some Dependency is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Some Dependency.\n\n"))
			})
		})

		context("when it is after the the deprecation date", func() {
			it("returns a warning that the version of the dependency is no longer supported", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate.Add(24 * time.Hour)

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Some Dependency",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Some Dependency version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Some Dependency is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Some Dependency.\n\n"))
			})
		})
	})

	context("Candidates", func() {
		it("logs the candidate entries", func() {
			emitter.Candidates([]packit.BuildpackPlanEntry{
				{
					Metadata: map[string]interface{}{
						"version-source": "some-source",
						"version":        "some-version",
					},
				},
				{
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			})
			Expect(buffer.String()).To(Equal(`    Candidate version sources (in priority order):
      some-source -> "some-version"
      <unknown>   -> "other-version"

`))
		})

		context("when there are deuplicate version sources with the same version", func() {
			it("logs the candidate entries and removes duplicates", func() {
				emitter.Candidates([]packit.BuildpackPlanEntry{
					{
						Metadata: map[string]interface{}{
							"version-source": "some-source",
							"version":        "some-version",
						},
					},
					{
						Metadata: map[string]interface{}{
							"version-source": "some-source",
							"version":        "other-version",
						},
					},
					{
						Metadata: map[string]interface{}{
							"version-source": "other-source",
							"version":        "some-version",
						},
					},
					{
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Metadata: map[string]interface{}{
							"version-source": "some-source",
							"version":        "some-version",
						},
					},
				})
				Expect(buffer.String()).To(Equal(`    Candidate version sources (in priority order):
      some-source  -> "some-version"
      some-source  -> "other-version"
      other-source -> "some-version"
      <unknown>    -> "other-version"

`), buffer.String())
			})
		})
	})
}
