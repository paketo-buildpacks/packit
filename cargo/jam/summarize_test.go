package main_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSummarize(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		buildpackage string
		buffer       *bytes.Buffer
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "buildpackage")
		Expect(err).NotTo(HaveOccurred())

		tw := tar.NewWriter(file)

		firstBuildpack := bytes.NewBuffer(nil)
		firstBuildpackGW := gzip.NewWriter(firstBuildpack)
		firstBuildpackTW := tar.NewWriter(firstBuildpackGW)

		content := []byte(`[buildpack]
id = "some-buildpack"
name = "Some Buildpack"
version = "1.2.3"

[metadata.default-versions]
some-dependency = "1.2.x"
other-dependency = "2.3.x"

[[metadata.dependencies]]
	id = "some-dependency"
	stacks = ["some-stack"]
	version = "1.2.3"
	sha256 = "some-sha"

[[metadata.dependencies]]
	id = "other-dependency"
	stacks = ["other-stack"]
	version = "2.3.4"
	sha256 = "other-sha"

[[stacks]]
	id = "some-stack"

[[stacks]]
	id = "other-stack"
`)

		err = firstBuildpackTW.WriteHeader(&tar.Header{
			Name: "./buildpack.toml",
			Mode: 0644,
			Size: int64(len(content)),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = firstBuildpackTW.Write(content)
		Expect(err).NotTo(HaveOccurred())

		Expect(firstBuildpackTW.Close()).To(Succeed())
		Expect(firstBuildpackGW.Close()).To(Succeed())

		err = tw.WriteHeader(&tar.Header{
			Name: "blobs/sha256/first-buildpack-sha",
			Mode: 0644,
			Size: int64(firstBuildpack.Len()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write(firstBuildpack.Bytes())
		Expect(err).NotTo(HaveOccurred())

		secondBuildpack := bytes.NewBuffer(nil)
		secondBuildpackGW := gzip.NewWriter(secondBuildpack)
		secondBuildpackTW := tar.NewWriter(secondBuildpackGW)

		content = []byte(`[buildpack]
id = "other-buildpack"
name = "Other Buildpack"
version = "2.3.4"

[metadata.default-versions]
first-dependency = "4.5.x"
second-dependency = "5.6.x"

[[metadata.dependencies]]
	id = "first-dependency"
	stacks = ["first-stack"]
	version = "4.5.6"
	sha256 = "first-sha"

[[metadata.dependencies]]
	id = "second-dependency"
	stacks = ["second-stack"]
	version = "5.6.7"
	sha256 = "second-sha"

[[stacks]]
	id = "first-stack"

[[stacks]]
	id = "second-stack"
`)

		err = secondBuildpackTW.WriteHeader(&tar.Header{
			Name: "./buildpack.toml",
			Mode: 0644,
			Size: int64(len(content)),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = secondBuildpackTW.Write(content)
		Expect(err).NotTo(HaveOccurred())

		Expect(secondBuildpackTW.Close()).To(Succeed())
		Expect(secondBuildpackGW.Close()).To(Succeed())

		err = tw.WriteHeader(&tar.Header{
			Name: "blobs/sha256/second-buildpack-sha",
			Mode: 0644,
			Size: int64(secondBuildpack.Len()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write(secondBuildpack.Bytes())
		Expect(err).NotTo(HaveOccurred())

		metaBuildpack := bytes.NewBuffer(nil)
		metaBuildpackGW := gzip.NewWriter(metaBuildpack)
		metaBuildpackTW := tar.NewWriter(metaBuildpackGW)

		content = []byte(`[buildpack]
id = "meta-buildpack"
name = "Meta Buildpack"
version = "3.4.5"

[[order]]
	[[order.group]]
	id = "some-buildpack"
	version = "1.2.3"

	[[order.group]]
	id = "other-buildpack"
	version = "2.3.4"
`)

		err = metaBuildpackTW.WriteHeader(&tar.Header{
			Name: "./buildpack.toml",
			Mode: 0644,
			Size: int64(len(content)),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = metaBuildpackTW.Write(content)
		Expect(err).NotTo(HaveOccurred())

		Expect(metaBuildpackTW.Close()).To(Succeed())
		Expect(metaBuildpackGW.Close()).To(Succeed())

		err = tw.WriteHeader(&tar.Header{
			Name: "blobs/sha256/meta-buildpack-sha",
			Mode: 0644,
			Size: int64(metaBuildpack.Len()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write(metaBuildpack.Bytes())
		Expect(err).NotTo(HaveOccurred())

		manifest := bytes.NewBuffer(nil)
		err = json.NewEncoder(manifest).Encode(map[string]interface{}{
			"layers": []map[string]interface{}{
				{"digest": "sha256:first-buildpack-sha"},
				{"digest": "sha256:second-buildpack-sha"},
				{"digest": "sha256:meta-buildpack-sha"},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = tw.WriteHeader(&tar.Header{
			Name: "blobs/sha256/manifest-sha",
			Mode: 0644,
			Size: int64(manifest.Len()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write(manifest.Bytes())
		Expect(err).NotTo(HaveOccurred())

		index := bytes.NewBuffer(nil)
		err = json.NewEncoder(index).Encode(map[string]interface{}{
			"manifests": []map[string]interface{}{
				{"digest": "sha256:manifest-sha"},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = tw.WriteHeader(&tar.Header{
			Name: "index.json",
			Mode: 0644,
			Size: int64(index.Len()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write(index.Bytes())
		Expect(err).NotTo(HaveOccurred())

		buildpackage = file.Name()

		Expect(tw.Close()).To(Succeed())
		Expect(file.Close()).To(Succeed())

		buffer = bytes.NewBuffer(nil)
	})

	it.After(func() {
		Expect(os.Remove(buildpackage)).To(Succeed())
	})

	context("when the format is set to markdown", func() {
		it("prints out the summary of a buildpack tarball", func() {
			command := exec.Command(
				path, "summarize",
				"--buildpack", buildpackage,
				"--format", "markdown",
			)
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), func() string { return buffer.String() })

			Expect(string(session.Out.Contents())).To(Equal(`## Meta Buildpack 3.4.5` +

				"\n\n**ID:** `meta-buildpack`\n\n" +

				"**Digest:** `sha256:manifest-sha`" +
				`

#### Included Buildpackages:
| Name | ID | Version |
|---|---|---|
| Some Buildpack | some-buildpack | 1.2.3 |
| Other Buildpack | other-buildpack | 2.3.4 |

<details>
<summary>Order Groupings</summary>

| ID | Version | Optional |
|---|---|---|
| some-buildpack | 1.2.3 | false |
| other-buildpack | 2.3.4 | false |

</details>

---

<details>
<summary>Some Buildpack 1.2.3</summary>` +

				"\n\n**ID:** `some-buildpack`\n\n" +

				`#### Supported Stacks:
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
| other-dependency | 2.3.4 | other-sha |
| some-dependency | 1.2.3 | some-sha |

---

</details>

<details>
<summary>Other Buildpack 2.3.4</summary>` +

				"\n\n**ID:** `other-buildpack`\n\n" +

				`#### Supported Stacks:
- first-stack
- second-stack

#### Default Dependency Versions:
| ID | Version |
|---|---|
| first-dependency | 4.5.x |
| second-dependency | 5.6.x |

#### Dependencies:
| Name | Version | SHA256 |
|---|---|---|
| first-dependency | 4.5.6 | first-sha |
| second-dependency | 5.6.7 | second-sha |

---

</details>
`))
		})
	})

	context("when the format is set to json", func() {
		it("prints out the summary of a buildpack tarball", func() {
			command := exec.Command(
				path, "summarize",
				"--buildpack", buildpackage,
				"--format", "json",
			)
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), func() string { return buffer.String() })

			Expect(buffer.String()).To(MatchJSON(`{
				"buildpackage": {
					"buildpack": {
						"id": "meta-buildpack",
						"name": "Meta Buildpack",
						"version": "3.4.5"
					},
					"metadata": {},
					"order": [{
						"group": [{
								"id": "some-buildpack",
								"version": "1.2.3"
							},
							{
								"id": "other-buildpack",
								"version": "2.3.4"
							}
						]
					}]
				},
				"children": [{
						"buildpack": {
							"id": "some-buildpack",
							"name": "Some Buildpack",
							"version": "1.2.3"
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
									"version": "1.2.3",
									"sha256": "some-sha"

								},
								{
									"id": "other-dependency",
									"stacks": [
										"other-stack"
									],
									"version": "2.3.4",
									"sha256": "other-sha"
								}
							]
						},
						"stacks": [{
							"id": "some-stack"
						},
						{
							"id": "other-stack"
						}]
					},
					{
						"buildpack": {
							"id": "other-buildpack",
							"name": "Other Buildpack",
							"version": "2.3.4"
						},
						"metadata": {
							"default-versions": {
								"first-dependency": "4.5.x",
								"second-dependency": "5.6.x"
							},
							"dependencies": [{
									"id": "first-dependency",
									"stacks": [
										"first-stack"
									],
									"version": "4.5.6",
									"sha256": "first-sha"
								},
								{
									"id": "second-dependency",
									"stacks": [
										"second-stack"
									],
									"version": "5.6.7",
									"sha256": "second-sha"
								}
							]
						},
						"stacks": [{
							"id": "first-stack"
						},
						{
							"id": "second-stack"
						}]
					}
				]
			}`))
		})
	})

	context("failure cases", func() {
		context("when the required buildpack flag is not set", func() {
			it("prints an error message", func() {
				command := exec.Command(
					path, "summarize",
					"--format", "markdown",
				)
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err.Contents()).To(ContainSubstring("Error: required flag(s) \"buildpack\" not set"))
			})
		})
	})
}
