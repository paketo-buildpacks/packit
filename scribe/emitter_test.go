package scribe_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
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
			entry      packit.BuildpackPlanEntry
			dependency postal.Dependency
		)

		it.Before(func() {
			now = time.Now()

			entry = packit.BuildpackPlanEntry{
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
				emitter.SelectedDependency(packit.BuildpackPlanEntry{}, dependency, now)
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

				entry = packit.BuildpackPlanEntry{
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

				entry = packit.BuildpackPlanEntry{
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

				entry = packit.BuildpackPlanEntry{
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

	context("WithLevel", func() {
		context("default", func() {
			it("output includes debug level logs", func() {
				emitter.Title("non-debug title")
				emitter.Debug.Title("debug title")

				Expect(buffer.String()).To(ContainLines(
					"non-debug title",
				))
				Expect(buffer.String()).ToNot(ContainLines(
					"debug title",
				))
			})
		})

		context("DEBUG", func() {
			it.Before(func() {
				emitter = emitter.WithLevel("DEBUG")
			})

			it("output includes debug level logs", func() {
				emitter.Title("non-debug title")
				emitter.Debug.Title("debug title")

				Expect(buffer.String()).To(ContainLines(
					"non-debug title",
					"debug title",
				))
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
			Expect(buffer.String()).To(ContainLines(
				"    Candidate version sources (in priority order):",
				`      some-source -> "some-version"`,
				`      <unknown>   -> "other-version"`,
				"",
			))
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
		var processes []packit.Process

		it.Before(func() {
			processes = []packit.Process{
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
					Type:    "some-other-type",
					Command: "some-other-command",
					Args:    []string{"some", "args"},
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
			var processEnvs []map[string]packit.Environment

			it.Before(func() {
				processEnvs = []map[string]packit.Environment{
					{
						"web": packit.Environment{
							"WEB_VAR.default": "some-env",
						},
					},
					{
						"web": packit.Environment{
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
			emitter.EnvironmentVariables(packit.Layer{
				BuildEnv: packit.Environment{
					"NODE_HOME.default":    "/some/path",
					"NODE_ENV.default":     "some-env",
					"NODE_VERBOSE.default": "some-bool",
				},
				LaunchEnv: packit.Environment{
					"NODE_HOME.default":    "/some/path",
					"NODE_ENV.default":     "another-env",
					"NODE_VERBOSE.default": "another-bool",
				},
				SharedEnv: packit.Environment{
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
				emitter.EnvironmentVariables(packit.Layer{
					LaunchEnv: packit.Environment{
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
				emitter.EnvironmentVariables(packit.Layer{
					BuildEnv: packit.Environment{
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
	context("LayerFlags", func() {
		context("when log level is INFO", func() {
			it("prints information about launch, cache, build flags on layer", func() {
				emitter.LayerFlags(packit.Layer{Name: "some-layer"})
				Expect(buffer.String()).To(BeEmpty())
			})
		})
		context("when log level is DEBUG", func() {
			it.Before(func() {
				emitter = scribe.NewEmitter(buffer).WithLevel("DEBUG")
			})
			it("prints information about launch, cache, build flags on layer", func() {
				emitter.LayerFlags(packit.Layer{
					Name:   "some-layer",
					Launch: false,
					Build:  true,
					Cache:  false,
				})
				Expect(buffer.String()).To(ContainLines(
					"  Setting up layer 'some-layer'",
					"    Available at app launch: false",
					"    Available to other buildpacks: true",
					"    Cached for rebuilds: false",
				))
			})
		})
	})

	context("GeneratingSBOM", func() {
		it("prints the correct log line with the inputted path", func() {
			emitter.GeneratingSBOM("/some/path")

			Expect(buffer.String()).To(ContainSubstring("Generating SBOM for /some/path"))
		})
	})

	context("FormattingSBOM", func() {
		context("when log level is INFO", func() {
			it("does not print anything", func() {
				emitter.FormattingSBOM("format1", "format2")
				Expect(buffer.String()).To(BeEmpty())
			})

			context("when the log level is DEBUG", func() {
				it.Before(func() {
					emitter = scribe.NewEmitter(buffer).WithLevel("DEBUG")
				})

				it("lists the inputted SBOM formats", func() {
					emitter.FormattingSBOM("format1", "format2")
					Expect(buffer.String()).To(ContainLines(
						"  Writing SBOM in the following format(s):",
						"    format1",
						"    format2",
					))
				})
			})
		})
	})

	context("BuildConfiguration", func() {
		context("when log level is INFO", func() {
			it("does not print anything", func() {
				emitter.BuildConfiguration(map[string]string{"ENV_VAR": "value"})
				Expect(buffer.String()).To(BeEmpty())
			})

			context("when the log level is DEBUG", func() {
				it.Before(func() {
					emitter = scribe.NewEmitter(buffer).WithLevel("DEBUG")
				})

				it("lists the environment variables of the build configuration and their values", func() {
					emitter.BuildConfiguration(map[string]string{
						"OTHER_ENV_VAR": "another-value",
						"ENV_VAR":       "value",
					})
					Expect(buffer.String()).To(ContainLines(
						"  Build configuration:",
						`    ENV_VAR       -> "value"`,
						`    OTHER_ENV_VAR -> "another-value"`,
					))
				})
			})
		})
	})
}
