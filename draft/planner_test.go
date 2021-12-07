package draft_test

import (
	"regexp"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanner(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		planner draft.Planner

		priorities []interface{}
	)

	it.Before(func() {
		priorities = []interface{}{
			"highest",
			"lowest",
		}

		planner = draft.NewPlanner()
	})

	context("ResolveEntries", func() {
		it("resolves the best plan entry", func() {
			entry, entries := planner.Resolve("node", []packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "another-version",
						"version-source": "lowest",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "npm",
					Metadata: map[string]interface{}{
						"version":        "some-version",
						"version-source": "highest",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "some-version",
						"version-source": "highest",
					},
				},
			}, priorities)

			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "some-version",
					"version-source": "highest",
				},
			}))

			Expect(entries).To(Equal([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "some-version",
						"version-source": "highest",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "another-version",
						"version-source": "lowest",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			}))
		})

		context("the priorities are nil", func() {
			it("returns the first entry in the filtered map", func() {
				entry, entries := planner.Resolve("node", []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "another-version",
							"version-source": "lowest",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Name: "npm",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
						},
					},
				}, nil)

				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "another-version",
						"version-source": "lowest",
					},
				}))

				Expect(entries).To(Equal([]packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "another-version",
							"version-source": "lowest",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
						},
					},
				}))
			})
		})

		context("there are no entries matching the given name", func() {
			it("returns no entries", func() {
				_, entries := planner.Resolve("some-name", []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "another-version",
							"version-source": "lowest",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Name: "npm",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
						},
					},
				}, priorities)

				Expect(entries).To(BeNil())
			})
		})

		context("there are entries matching a regexp", func() {
			it.Before(func() {
				priorities = []interface{}{
					"buildpack.yml",
					regexp.MustCompile(`^.*\.(cs)|(fs)|(vb)proj$`),
					regexp.MustCompile(`^.*\.runtimeconfig\.json$`),
				}
			})

			it("returns no entries", func() {
				entry, entries := planner.Resolve("dotnet-runtime", []packit.BuildpackPlanEntry{
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "myapp.runtimeconfig.json",
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "another-version",
							"version-source": "myapp.vbproj",
						},
					},
				}, priorities)

				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "dotnet-runtime",
					Metadata: map[string]interface{}{
						"version":        "another-version",
						"version-source": "myapp.vbproj",
					},
				}))
				Expect(entries).To(Equal([]packit.BuildpackPlanEntry{
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "another-version",
							"version-source": "myapp.vbproj",
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "myapp.runtimeconfig.json",
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
				}))
			})
		})
	})

	context("MergeLayerTypes", func() {
		it("resolves the layer types from plan metadata", func() {
			launch, build := planner.MergeLayerTypes("node", []packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"build": true,
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
						"launch":  true,
					},
				},
				{
					Name: "npm",
					Metadata: map[string]interface{}{
						"version":        "some-version",
						"version-source": "highest",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "some-version",
						"version-source": "highest",
					},
				},
			})

			Expect(launch).To(BeTrue())
			Expect(build).To(BeTrue())
		})

		context("if there are flags set in irrelevant entries", func() {
			it("resolves the layer types from plan metadata and ignores the irrelevant", func() {
				launch, build := planner.MergeLayerTypes("node", []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version": "other-version",
						},
					},
					{
						Name: "npm",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
							"launch":         true,
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "some-version",
							"version-source": "highest",
						},
					},
				})

				Expect(launch).To(BeFalse())
				Expect(build).To(BeTrue())
			})
		})
	})
}
