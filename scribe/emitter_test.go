package scribe_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
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
		var (
			now        time.Time
			entry      libcnb.BuildpackPlanEntry
			dependency postal.Dependency
		)

		it.Before(func() {
			now = time.Now()

			entry = libcnb.BuildpackPlanEntry{
				Metadata: map[string]interface{}{"version-source": "some-source"},
			}

			dependency = postal.Dependency{
				Name:    "Some Dependency",
				Version: "some-version",
			}
		})

		it("prints details about the selected dependency", func() {
			emitter.SelectedDependency(entry, dependency, now)
			Expect(buffer.String()).To(ContainLines(
				"    Selected Some Dependency version (using some-source): some-version",
				"",
			))
		})

		context("when the version source is missing", func() {
			it("prints details about the selected dependency", func() {
				emitter.SelectedDependency(libcnb.BuildpackPlanEntry{}, dependency, now)
				Expect(buffer.String()).To(ContainLines(
					"    Selected Some Dependency version (using <unknown>): some-version",
					"",
				))
			})
		})

		context("when it is within 30 days of the deprecation date", func() {
			it.Before(func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now = deprecationDate.Add(-29 * 24 * time.Hour)

				entry = libcnb.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency = postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Some Dependency",
					Version:         "some-version",
				}
			})

			it("returns a warning that the dependency will be deprecated after the deprecation date", func() {
				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainLines(
					"    Selected Some Dependency version (using some-source): some-version",
					"      Version some-version of Some Dependency will be deprecated after 2021-04-01.",
					"      Migrate your application to a supported version of Some Dependency before this time.",
					"",
				))
			})
		})

		context("when it is on the the deprecation date", func() {
			it.Before(func() {
				var err error
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now = deprecationDate

				entry = libcnb.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}

				dependency = postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Some Dependency",
					Version:         "some-version",
				}
			})

			it("returns a warning that the version of the dependency is no longer supported", func() {
				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainLines(
					"    Selected Some Dependency version (using some-source): some-version",
					"      Version some-version of Some Dependency is deprecated.",
					"      Migrate your application to a supported version of Some Dependency.",
					"",
				))
			})
		})

		context("when it is after the the deprecation date", func() {
			it.Before(func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now = deprecationDate.Add(24 * time.Hour)

				entry = libcnb.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency = postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Some Dependency",
					Version:         "some-version",
				}
			})

			it("returns a warning that the version of the dependency is no longer supported", func() {
				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainLines(
					"    Selected Some Dependency version (using some-source): some-version",
					"      Version some-version of Some Dependency is deprecated.",
					"      Migrate your application to a supported version of Some Dependency.",
					"",
				))
			})
		})
	})

	context("Candidates", func() {
		it("logs the candidate entries", func() {
			emitter.Candidates([]libcnb.BuildpackPlanEntry{
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
			Expect(buffer.String()).To(ContainLines(
				"    Candidate version sources (in priority order):",
				`      some-source -> "some-version"`,
				`      <unknown>   -> "other-version"`,
				"",
			))
		})

		context("when there are deuplicate version sources with the same version", func() {
			it("logs the candidate entries and removes duplicates", func() {
				emitter.Candidates([]libcnb.BuildpackPlanEntry{
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

				Expect(buffer.String()).To(ContainLines(
					"    Candidate version sources (in priority order):",
					`      some-source  -> "some-version"`,
					`      some-source  -> "other-version"`,
					`      other-source -> "some-version"`,
					`      <unknown>    -> "other-version"`,
					"",
				))
			})
		})
	})

	context("LaunchProcesses", func() {
		var processes []libcnb.Process

		it.Before(func() {
			processes = []libcnb.Process{
				{
					Type:    "some-type",
					Command: "some-command",
				},
				{
					Type:    "web",
					Command: "web-command",
					Default: true,
				},
				{
					Type:      "some-other-type",
					Command:   "some-other-command",
					Arguments: []string{"some", "args"},
				},
			}
		})

		it("prints a list of launch processes", func() {
			emitter.LaunchProcesses(processes)

			Expect(buffer.String()).To(ContainLines(
				"  Assigning launch processes:",
				"    some-type:       some-command",
				"    web (default):   web-command",
				"    some-other-type: some-other-command some args",
				"",
			))
		})

		context("when passed process specific environment variables", func() {
			var processEnvs []map[string]libcnb.Environment

			it.Before(func() {
				processEnvs = []map[string]libcnb.Environment{
					{
						"web": libcnb.Environment{
							"WEB_VAR.default": "some-env",
						},
					},
					{
						"web": libcnb.Environment{
							"ANOTHER_WEB_VAR.default": "another-env",
						},
					},
				}
			})

			it("prints a list of the launch processes and their processes specific env vars", func() {
				emitter.LaunchProcesses(processes, processEnvs...)

				Expect(buffer.String()).To(ContainLines(
					"  Assigning launch processes:",
					"    some-type:       some-command",
					"    web (default):   web-command",
					`      ANOTHER_WEB_VAR -> "another-env"`,
					`      WEB_VAR         -> "some-env"`,
					"    some-other-type: some-other-command some args",
					"",
				))
			})
		})
	})

	context("EnvironmentVariables", func() {
		it("prints a list of environment variables available during launch and build", func() {
			emitter.EnvironmentVariables(libcnb.Layer{
				BuildEnvironment: libcnb.Environment{
					"NODE_HOME.default":    "/some/path",
					"NODE_ENV.default":     "some-env",
					"NODE_VERBOSE.default": "some-bool",
				},
				LaunchEnvironment: libcnb.Environment{
					"NODE_HOME.default":    "/some/path",
					"NODE_ENV.default":     "another-env",
					"NODE_VERBOSE.default": "another-bool",
				},
				SharedEnvironment: libcnb.Environment{
					"SHARED_ENV.default": "shared-env",
				},
			})

			Expect(buffer.String()).To(ContainLines(
				"  Configuring build environment",
				`    NODE_ENV     -> "some-env"`,
				`    NODE_HOME    -> "/some/path"`,
				`    NODE_VERBOSE -> "some-bool"`,
				`    SHARED_ENV   -> "shared-env"`,
				"",
				"  Configuring launch environment",
				`    NODE_ENV     -> "another-env"`,
				`    NODE_HOME    -> "/some/path"`,
				`    NODE_VERBOSE -> "another-bool"`,
				`    SHARED_ENV   -> "shared-env"`,
				"",
			))
		})

		context("when one of the environments is empty it only prints the one that has set vars", func() {
			it("prints a list of environment variables available during launch", func() {
				emitter.EnvironmentVariables(libcnb.Layer{
					LaunchEnvironment: libcnb.Environment{
						"NODE_HOME.default":    "/some/path",
						"NODE_ENV.default":     "another-env",
						"NODE_VERBOSE.default": "another-bool",
					},
				})

				Expect(buffer.String()).To(ContainLines(
					"  Configuring launch environment",
					`    NODE_ENV     -> "another-env"`,
					`    NODE_HOME    -> "/some/path"`,
					`    NODE_VERBOSE -> "another-bool"`,
					"",
				))

				Expect(buffer.String()).NotTo(ContainSubstring("  Configuring build environment"))
			})

			it("prints a list of environment variables available during build", func() {
				emitter.EnvironmentVariables(libcnb.Layer{
					BuildEnvironment: libcnb.Environment{
						"NODE_HOME.default":    "/some/path",
						"NODE_ENV.default":     "some-env",
						"NODE_VERBOSE.default": "some-bool",
					},
				})

				Expect(buffer.String()).To(ContainLines(
					"  Configuring build environment",
					`    NODE_ENV     -> "some-env"`,
					`    NODE_HOME    -> "/some/path"`,
					`    NODE_VERBOSE -> "some-bool"`,
					"",
				))

				Expect(buffer.String()).NotTo(ContainSubstring("  Configuring launch environment"))
			})
		})
	})
}
