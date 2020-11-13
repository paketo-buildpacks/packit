package judge_test

import (
	"bytes"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/judge"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryHandler(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		resolver judge.PlanEntryHandler

		buffer *bytes.Buffer

		priorities map[string]int
	)

	it.Before(func() {
		priorities = map[string]int{
			"highest": 2,
			"lowest":  1,
			"":        -1,
		}

		buffer = bytes.NewBuffer(nil)

		resolver = judge.NewPlanEntryHandler(scribe.NewLogger(buffer))
	})

	context("ResolveEntries", func() {
		it("resolves the best plan entry", func() {
			entry, ok := resolver.ResolveEntries("node", []packit.BuildpackPlanEntry{
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

			Expect(ok).To(BeTrue())
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "some-version",
					"version-source": "highest",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      highest   -> \"some-version\""))
			Expect(buffer.String()).To(ContainSubstring("      lowest    -> \"another-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown> -> \"other-version\""))
		})

		context("the priorities are nil", func() {
			it("returns the first entry in the filtered map", func() {
				entry, ok := resolver.ResolveEntries("node", []packit.BuildpackPlanEntry{
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

				Expect(ok).To(BeTrue())
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "another-version",
						"version-source": "lowest",
					},
				}))
			})
		})

		context("there are no enties matching the given name", func() {
			it("returns the first entry in the filtered map", func() {
				_, ok := resolver.ResolveEntries("some-name", []packit.BuildpackPlanEntry{
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

				Expect(ok).To(BeFalse())
			})
		})
	})

	context("MergeLayerTypes", func() {
		it("resolves the layer types from plan metadata", func() {
			layerTypes := resolver.MergeLayerTypes("node", []packit.BuildpackPlanEntry{
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

			Expect(layerTypes).To(ConsistOf([]packit.LayerType{
				packit.LaunchLayer,
				packit.BuildLayer,
				packit.CacheLayer,
			}))
		})

		context("if there are flags set in irrelevant entries", func() {
			it("resolves the layer types from plan metadata and ignores the irrelevant", func() {
				layerTypes := resolver.MergeLayerTypes("node", []packit.BuildpackPlanEntry{
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

				Expect(layerTypes).NotTo(ConsistOf([]packit.LayerType{packit.LaunchLayer}))
			})
		})
	})
}
