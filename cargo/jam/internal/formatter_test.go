package internal_test

import (
	"bytes"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormatter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer    *bytes.Buffer
		formatter internal.Formatter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		formatter = internal.NewFormatter(buffer)
	})

	context("Markdown", func() {
		it("returns a list of dependencies", func() {
			formatter.Markdown([]cargo.Config{
				{
					Buildpack: cargo.ConfigBuildpack{
						ID:      "some-buildpack",
						Version: "some-version",
					},
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID:      "some-dependency",
								Stacks:  []string{"some-stack"},
								Version: "1.2.3",
							},
							{
								ID:      "some-dependency",
								Stacks:  []string{"other-stack"},
								Version: "1.2.3",
							},
							{
								ID:      "other-dependency",
								Stacks:  []string{"some-stack", "other-stack"},
								Version: "2.3.4",
							},
							{
								ID:      "other-dependency",
								Stacks:  []string{"other-stack"},
								Version: "2.3.5",
							},
						},
						DefaultVersions: map[string]string{
							"some-dependency":  "1.2.x",
							"other-dependency": "2.3.x",
						},
					},
					Stacks: []cargo.ConfigStack{
						{ID: "some-stack"},
						{ID: "other-stack"},
					},
				},
			})
			Expect(buffer.String()).To(Equal(`## some-buildpack some-version
### Dependencies
| name | version | stacks |
|---|---|---|
| other-dependency | 2.3.5 | other-stack |
| other-dependency | 2.3.4 | other-stack, some-stack |
| some-dependency | 1.2.3 | other-stack, some-stack |

### Default Dependencies
| name | version |
|---|---|
| other-dependency | 2.3.x |
| some-dependency | 1.2.x |

### Supported Stacks
| name |
|---|
| other-stack |
| some-stack |
`))
		})

		context("when dependencies and default-versions are empty", func() {
			it("returns a list of dependencies", func() {
				formatter.Markdown([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Version: "some-version",
						},
						Stacks: []cargo.ConfigStack{
							{ID: "some-stack"},
							{ID: "other-stack"},
						},
					},
				})
				Expect(buffer.String()).To(Equal(`## some-buildpack some-version
### Supported Stacks
| name |
|---|
| other-stack |
| some-stack |
`))
			})
		})

		context("when stacks are empty", func() {
			it("returns a list of dependencies", func() {
				formatter.Markdown([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Version: "some-version",
						},
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{
									ID:      "some-dependency",
									Stacks:  []string{"some-stack"},
									Version: "1.2.3",
								},
								{
									ID:      "some-dependency",
									Stacks:  []string{"other-stack"},
									Version: "1.2.3",
								},
								{
									ID:      "other-dependency",
									Stacks:  []string{"some-stack", "other-stack"},
									Version: "2.3.4",
								},
								{
									ID:      "other-dependency",
									Stacks:  []string{"other-stack"},
									Version: "2.3.5",
								},
							},
							DefaultVersions: map[string]string{
								"some-dependency":  "1.2.x",
								"other-dependency": "2.3.x",
							},
						},
					},
				})
				Expect(buffer.String()).To(Equal(`## some-buildpack some-version
### Dependencies
| name | version | stacks |
|---|---|---|
| other-dependency | 2.3.5 | other-stack |
| other-dependency | 2.3.4 | other-stack, some-stack |
| some-dependency | 1.2.3 | other-stack, some-stack |

### Default Dependencies
| name | version |
|---|---|
| other-dependency | 2.3.x |
| some-dependency | 1.2.x |

`))
			})
		})

		context("when there are order groupings", func() {
			it("prints a the order groupings", func() {
				formatter.Markdown([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Version: "some-version",
						},
						Order: []cargo.ConfigOrder{
							{
								Group: []cargo.ConfigOrderGroup{
									{
										ID:      "some-buildpack",
										Version: "1.2.3",
									},
									{
										ID:       "optional-buildpack",
										Version:  "2.3.4",
										Optional: true,
									},
								},
							},
							{
								Group: []cargo.ConfigOrderGroup{
									{
										ID:      "other-buildpack",
										Version: "3.4.5",
									},
								},
							},
						},
					},
				})
				Expect(buffer.String()).To(Equal(`# some-buildpack some-version
### Order Groupings
| name | version | optional |
|---|---|---|
| some-buildpack | 1.2.3 | false |
| optional-buildpack | 2.3.4 | true |

| name | version | optional |
|---|---|---|
| other-buildpack | 3.4.5 | false |

`))
			})
		})
	})

	context("JSON", func() {
		it("returns a list of dependencies", func() {
			formatter.JSON([]cargo.Config{
				{
					Buildpack: cargo.ConfigBuildpack{
						ID:      "some-buildpack",
						Version: "some-version",
					},
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID:      "some-dependency",
								Stacks:  []string{"some-stack"},
								Version: "1.2.3",
							},
							{
								ID:      "some-dependency",
								Stacks:  []string{"other-stack"},
								Version: "1.2.3",
							},
							{
								ID:      "other-dependency",
								Stacks:  []string{"some-stack", "other-stack"},
								Version: "2.3.4",
							},
							{
								ID:      "other-dependency",
								Stacks:  []string{"other-stack"},
								Version: "2.3.5",
							},
						},
						DefaultVersions: map[string]string{
							"some-dependency":  "1.2.x",
							"other-dependency": "2.3.x",
						},
					},
					Stacks: []cargo.ConfigStack{
						{ID: "some-stack"},
						{ID: "other-stack"},
					},
				},
			})
			Expect(buffer.String()).To(MatchJSON(`{
	"buildpackage": {
		"buildpack": {
			"id": "some-buildpack",
			"version": "some-version"
		},
		"metadata": {
			"default-versions": {
				"some-dependency": "1.2.x",
				"other-dependency": "2.3.x"
			},
			"dependencies": [{
				"id": "some-dependency",
				"stacks": [
					"some-stack"
				],
				"version": "1.2.3"
			}, {
				"id": "some-dependency",
				"stacks": [
					"other-stack"
				],
				"version": "1.2.3"
			}, {
				"id": "other-dependency",
				"stacks": [
					"some-stack",
					"other-stack"
				],
				"version": "2.3.4"
			}, {
				"id": "other-dependency",
				"stacks": [
					"other-stack"
				],
				"version": "2.3.5"
			}]
		},
		"stacks": [{
				"id": "some-stack"
			},
			{
				"id": "other-stack"
			}
		]
	}
}`))
		})

		context("when buildpackage is a meta buildpackage", func() {
			it("prints the meta buildpackage info and the info of all its children", func() {
				formatter.JSON([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Version: "some-version",
						},
						Order: []cargo.ConfigOrder{
							{
								Group: []cargo.ConfigOrderGroup{
									{
										ID:      "some-buildpack",
										Version: "1.2.3",
									},
									{
										ID:       "optional-buildpack",
										Version:  "2.3.4",
										Optional: true,
									},
								},
							},
							{
								Group: []cargo.ConfigOrderGroup{
									{
										ID:      "other-buildpack",
										Version: "3.4.5",
									},
								},
							},
						},
					},
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Version: "some-version",
						},
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{
									ID:      "some-dependency",
									Stacks:  []string{"some-stack"},
									Version: "1.2.3",
								},
							},
							DefaultVersions: map[string]string{
								"some-dependency": "1.2.x",
							},
						},
						Stacks: []cargo.ConfigStack{
							{ID: "some-stack"},
						},
					},
				})
				Expect(buffer.String()).To(MatchJSON(`{
	"buildpackage": {
		"buildpack": {
			"id": "some-buildpack",
			"version": "some-version"
		},
		"metadata": {},
		"order": [{
				"group": [{
						"id": "some-buildpack",
						"version": "1.2.3"
					},
					{
						"id": "optional-buildpack",
						"version": "2.3.4",
						"optional": true
					}
				]

			},
			{
				"group": [{
					"id": "other-buildpack",
					"version": "3.4.5"
				}]
			}
		]
	},
	"children": [{
		"buildpack": {
			"id": "some-buildpack",
			"version": "some-version"
		},
		"metadata": {
			"default-versions": {
				"some-dependency": "1.2.x"
			},
			"dependencies": [{
				"id": "some-dependency",
				"stacks": [
					"some-stack"
				],
				"version": "1.2.3"
			}]
		},
		"stacks": [{
			"id": "some-stack"
		}]
	}]
}`))
			})
		})
	})
}
