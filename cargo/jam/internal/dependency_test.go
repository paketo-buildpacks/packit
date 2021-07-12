package internal_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDependency(t *testing.T, context spec.G, it spec.S) {
	var (
		withT           = NewWithT(t)
		Expect          = withT.Expect
		allDependencies []internal.Dependency
	)

	it.Before(func() {
		allDependencies = []internal.Dependency{
			{
				DeprecationDate: "",
				ID:              "some-dep",
				SHA256:          "some-sha",
				Source:          "some-source",
				SourceSHA256:    "some-source-sha",
				Stacks: []internal.Stack{
					{
						ID: "some-stack",
					},
				},
				URI:       "some-dep-uri",
				Version:   "v1.0.0",
				CreatedAt: "sometime",
				ModifedAt: "another-time",
				CPE:       "cpe-notation",
			},
			{
				DeprecationDate: "",
				ID:              "some-dep",
				SHA256:          "some-sha-two",
				Source:          "some-source-two",
				SourceSHA256:    "some-source-sha-two",
				Stacks: []internal.Stack{
					{
						ID: "some-stack-two",
					},
				},
				URI:       "some-dep-uri-two",
				Version:   "v1.1.2",
				CreatedAt: "sometime",
				ModifedAt: "another-time",
				CPE:       "cpe-notation",
			},
			{
				DeprecationDate: "",
				ID:              "some-dep",
				SHA256:          "some-sha-three",
				Source:          "some-source-three",
				SourceSHA256:    "some-source-sha-three",
				Stacks: []internal.Stack{
					{
						ID: "some-stack-three",
					},
				},
				URI:       "some-dep-uri-three",
				Version:   "v1.5.6",
				CreatedAt: "sometime",
				ModifedAt: "another-time",
				CPE:       "cpe-notation",
			},
			{
				DeprecationDate: "",
				ID:              "some-dep",
				SHA256:          "some-sha-four",
				Source:          "some-source-four",
				SourceSHA256:    "some-source-sha-four",
				Stacks: []internal.Stack{
					{
						ID: "some-stack-four",
					},
				},
				URI:       "some-dep-uri-four",
				Version:   "v2.3.2",
				CreatedAt: "sometime",
				ModifedAt: "another-time",
				CPE:       "cpe-notation",
			},
			{
				DeprecationDate: "",
				ID:              "different-dep",
				SHA256:          "different-dep-sha",
				Source:          "different-dep-source",
				SourceSHA256:    "different-dep-source-sha",
				Stacks: []internal.Stack{
					{
						ID: "different-dep-stack",
					},
				},
				URI:       "different-dep-uri",
				Version:   "v1.9.8",
				CreatedAt: "sometime",
				ModifedAt: "another-time",
				CPE:       "cpe-notation",
			},
			{
				DeprecationDate: "",
				ID:              "non-semver-dep",
				SHA256:          "non-semver-sha",
				Source:          "non-semver-source",
				SourceSHA256:    "non-semver-source-sha",
				Stacks: []internal.Stack{
					{
						ID: "non-semver-stack",
					},
				},
				URI:       "non-semver-uri",
				Version:   "non-semver1.9.8",
				CreatedAt: "sometime",
				ModifedAt: "another-time",
				CPE:       "cpe-notation",
			},
		}
	})

	context("GetAllDependencies", func() {
		var server *httptest.Server
		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				switch req.URL.Path {
				case "/v1/":
					w.WriteHeader(http.StatusOK)

				case "/v1/dependency":
					if req.URL.RawQuery == "name=some-dep" {
						w.WriteHeader(http.StatusOK)
						fmt.Fprintln(w, `[
  {
    "name": "some-dep",
    "version": "v1.0.0",
    "sha256": "some-sha",
    "uri": "some-dep-uri",
    "stacks": [
      {
        "id": "some-stack"
      }
    ],
    "source": "some-source",
    "source_sha256": "some-source-sha",
    "created_at": "sometime",
    "modified_at": "another-time",
		"cpe": "cpe-notation",
		"deprecation_date" : ""
  },
  {
    "name": "some-dep",
    "version": "v1.1.2",
    "sha256": "some-sha-two",
    "uri": "some-dep-uri-two",
    "stacks": [
      {
        "id": "some-stack-two"
      }
    ],
    "source": "some-source-two",
    "source_sha256": "some-source-sha-two",
    "created_at": "sometime",
    "modified_at": "another-time",
		"cpe": "cpe-notation"
  },
  {
    "name": "some-dep",
    "version": "v1.5.6",
    "sha256": "some-sha-three",
    "uri": "some-dep-uri-three",
    "stacks": [
      {
        "id": "some-stack-three"
      }
    ],
    "source": "some-source-three",
    "source_sha256": "some-source-sha-three",
    "created_at": "sometime",
    "modified_at": "another-time",
		"cpe": "cpe-notation"
  },
  {
    "name": "some-dep",
    "version": "v2.3.2",
    "sha256": "some-sha-four",
    "uri": "some-dep-uri-four",
    "stacks": [
      {
        "id": "some-stack-four"
      }
    ],
    "source": "some-source-four",
		"source_sha256": "some-source-sha-four",
    "created_at": "sometime",
    "modified_at": "another-time",
		"cpe": "cpe-notation"
  },
  {
    "name": "different-dep",
    "version": "v1.9.8",
    "sha256": "different-dep-sha",
    "uri": "different-dep-uri",
    "stacks": [
      {
        "id": "different-dep-stack"
      }
    ],
    "source": "different-dep-source",
		"source_sha256": "different-dep-source-sha",
    "created_at": "sometime",
    "modified_at": "another-time",
		"cpe": "cpe-notation"
  },
  {
    "name": "non-semver-dep",
    "version": "non-semver1.9.8",
    "sha256": "non-semver-sha",
    "uri": "non-semver-uri",
    "stacks": [
      {
        "id": "non-semver-stack"
      }
    ],
    "source": "non-semver-source",
		"source_sha256": "non-semver-source-sha",
    "created_at": "sometime",
    "modified_at": "another-time",
		"cpe": "cpe-notation"
  }
]`)
					}
					if req.URL.RawQuery == "name=bad-status" {
						w.WriteHeader(http.StatusBadRequest)
					}

					if req.URL.RawQuery == "name=bad-dep" {
						w.WriteHeader(http.StatusOK)
						filename := "other-payload"
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
					}

				default:
					t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
				}
			}))
		})

		it.After(func() {
			server.Close()
		})

		context("given a valid API and dependencyID", func() {
			it("returns a slice of Dependency from the API server", func() {
				dependencies, err := internal.GetAllDependencies(server.URL, "some-dep")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependencies).To(Equal(allDependencies))
			})
		})

		context("failure cases", func() {
			context("the url cannot be queried", func() {
				it("returns an error", func() {
					_, err := internal.GetAllDependencies("%%%", "some-dep")
					Expect(err).To(MatchError(ContainSubstring("failed to query url")))
				})
			})

			context("the API returns a non 200 status code", func() {
				it("returns an error", func() {
					_, err := internal.GetAllDependencies(server.URL, "bad-status")
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("failed to query url %s/v1/dependency?name=bad-status with: status code 400", server.URL))))
				})
			})

			context("the API response is not unmarshal-able", func() {
				it("returns an error", func() {
					_, err := internal.GetAllDependencies(server.URL, "bad-dep")
					Expect(err).To(MatchError(ContainSubstring("failed to unmarshal: unexpected end of JSON input")))
				})
			})
		})
	})

	context("GetDependenciesWithinConstraint", func() {
		context("given a valid api and constraint", func() {
			it("returns a sorted list of dependencies that match the constraint", func() {
				constraint := cargo.ConfigMetadataDependencyConstraint{
					Constraint: "1.*",
					ID:         "some-dep",
					Patches:    3,
				}

				dependencies, err := internal.GetDependenciesWithinConstraint(allDependencies, constraint, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependencies).To(Equal([]cargo.ConfigMetadataDependency{
					{
						CPE:          "cpe-notation",
						ID:           "some-dep",
						Version:      "1.0.0",
						Stacks:       []string{"some-stack"},
						URI:          "some-dep-uri",
						SHA256:       "some-sha",
						Source:       "some-source",
						SourceSHA256: "some-source-sha",
					},
					{
						CPE:          "cpe-notation",
						ID:           "some-dep",
						Version:      "1.1.2",
						Stacks:       []string{"some-stack-two"},
						URI:          "some-dep-uri-two",
						SHA256:       "some-sha-two",
						Source:       "some-source-two",
						SourceSHA256: "some-source-sha-two",
					},
					{
						CPE:          "cpe-notation",
						ID:           "some-dep",
						Version:      "1.5.6",
						Stacks:       []string{"some-stack-three"},
						URI:          "some-dep-uri-three",
						SHA256:       "some-sha-three",
						Source:       "some-source-three",
						SourceSHA256: "some-source-sha-three",
					},
				}))
			})
		})

		context("given a valid api and constraint that returns a dependency with non-semver version", func() {
			it("returns dependencies and makes the version semver-compatible", func() {
				constraint := cargo.ConfigMetadataDependencyConstraint{
					Constraint: "1.*",
					ID:         "non-semver-dep",
					Patches:    1,
				}

				dependencies, err := internal.GetDependenciesWithinConstraint(allDependencies, constraint, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependencies).To(Equal([]cargo.ConfigMetadataDependency{
					{
						CPE:          "cpe-notation",
						ID:           "non-semver-dep",
						Version:      "1.9.8",
						Stacks:       []string{"non-semver-stack"},
						URI:          "non-semver-uri",
						SHA256:       "non-semver-sha",
						Source:       "non-semver-source",
						SourceSHA256: "non-semver-source-sha",
					},
				}))
			})
		})

		context("failure cases", func() {
			context("given an invalid constraint", func() {
				it("returns an error", func() {
					constraint := cargo.ConfigMetadataDependencyConstraint{
						Constraint: "abc",
						ID:         "some-dep",
						Patches:    3,
					}

					_, err := internal.GetDependenciesWithinConstraint(allDependencies, constraint, "")
					Expect(err).To(MatchError("improper constraint: abc"))
				})
			})

			context("given a malformed dependency version", func() {
				it("returns an error", func() {
					constraint := cargo.ConfigMetadataDependencyConstraint{
						Constraint: "1.*",
						ID:         "some-dep",
						Patches:    3,
					}
					dependencies := []internal.Dependency{
						{
							DeprecationDate: "",
							ID:              "some-dep",
							SHA256:          "some-sha",
							Source:          "some-source",
							SourceSHA256:    "some-source-sha",
							Stacks: []internal.Stack{
								{
									ID: "some-stack",
								},
							},
							URI:       "some-dep-uri",
							Version:   "v1.xx",
							CreatedAt: "sometime",
							ModifedAt: "another-time",
							CPE:       "cpe-notation",
						},
					}

					_, err := internal.GetDependenciesWithinConstraint(dependencies, constraint, "")
					Expect(err).To(MatchError("Invalid Semantic Version"))
				})
			})
		})
	})

	context("FindDependencyName", func() {
		var cargoConfig cargo.Config
		it.Before(func() {
			cargoConfig = cargo.Config{
				API: "0.2",
				Buildpack: cargo.ConfigBuildpack{
					ID:       "some-buildpack-id",
					Name:     "some-buildpack-name",
					Version:  "some-buildpack-version",
					Homepage: "some-homepage-link",
				},
				Metadata: cargo.ConfigMetadata{
					Dependencies: []cargo.ConfigMetadataDependency{
						{
							ID:      "some-dependency",
							Name:    "Some Dependency Name",
							URI:     "http://some-url",
							Version: "1.2.3",
						},
					},
				},
			}
		})

		context("given a dependency ID and valid cargo.Config that contain that dependency", func() {
			it("returns the name of the dependency from the Config", func() {
				name := internal.FindDependencyName("some-dependency", cargoConfig)
				Expect(name).To(Equal("Some Dependency Name"))
			})
		})

		context("given a dependency ID and a cargo.Config that does not contain that dependency", func() {
			it("returns the empty string", func() {
				name := internal.FindDependencyName("unmatched-dependency", cargoConfig)
				Expect(name).To(Equal(""))
			})
		})
	})
}
