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
						Name:    "Some Buildpack",
						Version: "some-version",
						SHA256:  "sha256:some-buildpack-sha",
					},
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID:           "some-dependency",
								Stacks:       []string{"some-stack"},
								Version:      "1.2.3",
								SHA256:       "one-more-sha",
								Source:       "some-source",
								SourceSHA256: "source-sha",
							},
							{
								ID:      "some-dependency",
								Stacks:  []string{"other-stack"},
								Version: "1.2.3",
								SHA256:  "one-more-sha",
							},
							{
								ID:      "other-dependency",
								Stacks:  []string{"some-stack", "other-stack"},
								Version: "2.3.4",
								SHA256:  "another-sha",
							},
							{
								ID:      "other-dependency",
								Stacks:  []string{"other-stack"},
								Version: "2.3.5",
								SHA256:  "some-sha",
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
			Expect(buffer.String()).To(Equal(`## Some Buildpack some-version` +

				"\n\n**ID:** `some-buildpack`\n\n" +

				"**Digest:** `sha256:some-buildpack-sha`" +

				`

#### Supported Stacks:
- other-stack
- some-stack

#### Default Dependency Versions:
| ID | Version |
|---|---|
| other-dependency | 2.3.x |
| some-dependency | 1.2.x |

#### Dependencies:
| Name | Version | SHA256 |
|---|---|---|
| other-dependency | 2.3.5 | some-sha |
| other-dependency | 2.3.4 | another-sha |
| some-dependency | 1.2.3 | one-more-sha |

`))
		})

		context("when dependencies and default-versions are empty", func() {
			it("returns a list of dependencies", func() {
				formatter.Markdown([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Name:    "Some Buildpack",
							Version: "some-version",
							SHA256:  "sha256:some-buildpack-sha",
						},
						Stacks: []cargo.ConfigStack{
							{ID: "some-stack"},
							{ID: "other-stack"},
						},
					},
				})
				Expect(buffer.String()).To(Equal(`## Some Buildpack some-version` +

					"\n\n**ID:** `some-buildpack`\n\n" +

					"**Digest:** `sha256:some-buildpack-sha`" +

					`

#### Supported Stacks:
- other-stack
- some-stack

`))
			})
		})

		context("when stacks are empty", func() {
			it("returns a list of dependencies", func() {
				formatter.Markdown([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Name:    "Some Buildpack",
							Version: "some-version",
							SHA256:  "sha256:some-buildpack-sha",
						},
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{
									ID:      "some-dependency",
									Stacks:  []string{"some-stack"},
									Version: "1.2.3",
									SHA256:  "one-more-sha",
								},
								{
									ID:      "some-dependency",
									Stacks:  []string{"other-stack"},
									Version: "1.2.3",
									SHA256:  "one-more-sha",
								},
								{
									ID:      "other-dependency",
									Stacks:  []string{"some-stack", "other-stack"},
									Version: "2.3.4",
									SHA256:  "another-sha",
								},
								{
									ID:      "other-dependency",
									Stacks:  []string{"other-stack"},
									Version: "2.3.5",
									SHA256:  "some-sha",
								},
							},
							DefaultVersions: map[string]string{
								"some-dependency":  "1.2.x",
								"other-dependency": "2.3.x",
							},
						},
					},
				})
				Expect(buffer.String()).To(Equal(`## Some Buildpack some-version` +

					"\n\n**ID:** `some-buildpack`\n\n" +

					"**Digest:** `sha256:some-buildpack-sha`" +

					`

#### Default Dependency Versions:
| ID | Version |
|---|---|
| other-dependency | 2.3.x |
| some-dependency | 1.2.x |

#### Dependencies:
| Name | Version | SHA256 |
|---|---|---|
| other-dependency | 2.3.5 | some-sha |
| other-dependency | 2.3.4 | another-sha |
| some-dependency | 1.2.3 | one-more-sha |

`))
			})
		})

		context("when there are order groupings", func() {
			it("prints the order groupings", func() {
				formatter.Markdown([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "order-buildpack",
							Name:    "Order Buildpack",
							Version: "order-version",
							SHA256:  "sha256:order-buildpack-sha",
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
										ID:      "some-buildpack",
										Version: "1.2.3",
									},
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
							Name:    "Some Buildpack",
							Version: "1.2.3",
						},
					},
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "optional-buildpack",
							Name:    "Optional Buildpack",
							Version: "2.3.4",
						},
					},
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "other-buildpack",
							Name:    "Other Buildpack",
							Version: "3.4.5",
						},
					},
				})
				Expect(buffer.String()).To(Equal(`## Order Buildpack order-version` +

					"\n\n**ID:** `order-buildpack`\n\n" +

					"**Digest:** `sha256:order-buildpack-sha`" +

					`

#### Included Buildpackages:
| Name | ID | Version |
|---|---|---|
| Some Buildpack | some-buildpack | 1.2.3 |
| Optional Buildpack | optional-buildpack | 2.3.4 |
| Other Buildpack | other-buildpack | 3.4.5 |

<details>
<summary>Order Groupings</summary>

| ID | Version | Optional |
|---|---|---|
| some-buildpack | 1.2.3 | false |
| optional-buildpack | 2.3.4 | true |

| ID | Version | Optional |
|---|---|---|
| some-buildpack | 1.2.3 | false |
| other-buildpack | 3.4.5 | false |

</details>

---

<details>
<summary>Some Buildpack 1.2.3</summary>` +

					"\n\n**ID:** `some-buildpack`\n\n" +

					`---

</details>

<details>
<summary>Optional Buildpack 2.3.4</summary>` +

					"\n\n**ID:** `optional-buildpack`\n\n" +

					`---

</details>

<details>
<summary>Other Buildpack 3.4.5</summary>` +

					"\n\n**ID:** `other-buildpack`\n\n" +

					`---

</details>
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
