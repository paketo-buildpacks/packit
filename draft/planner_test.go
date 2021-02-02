package draft_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanner(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		planner draft.Planner

		priorities map[string]int
	)

	it.Before(func() {
		priorities = map[string]int{
			"highest": 2,
			"lowest":  1,
			"":        -1,
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

		context("there are no enties matching the given name", func() {
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
