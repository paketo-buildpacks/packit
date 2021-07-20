package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testUpdateDependencies(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		server       *httptest.Server
		buildpackDir string
	)

	it.Before(func() {
		var err error
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)
				return
			}

			switch req.URL.Path {
			case "/v2/":
				w.WriteHeader(http.StatusOK)

			case "/v1/dependency":
				if req.URL.RawQuery == "name=node" {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `[
  {
    "name": "node",
		"version": "v1.2.4",
    "sha256": "some-sha",
    "uri": "some-dep-uri",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      }
    ],
    "source": "some-source",
    "source_sha256": "some-source-sha",
		"cpe": "node-cpe",
		"purl": "some-purl",
		"licenses": ["MIT", "MIT-2"]
	},
  {
    "name": "node",
		"version": "v1.3.5",
    "sha256": "some-sha",
    "uri": "some-dep-uri",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      }
    ],
    "source": "some-source",
    "source_sha256": "some-source-sha",
		"cpe": "node-cpe",
		"purl": "some-purl",
		"licenses": ["MIT", "MIT-2"]
	},
  {
    "name": "node",
		"version": "v2.1.9",
    "sha256": "some-sha",
    "uri": "some-dep-uri",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      }
    ],
    "source": "some-source",
    "source_sha256": "some-source-sha",
		"cpe": "node-cpe",
		"purl": "some-purl",
		"licenses": ["MIT", "MIT-2"]
	},
  {
    "name": "node",
		"version": "v2.2.5",
    "sha256": "some-sha",
    "uri": "some-dep-uri",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      }
    ],
    "source": "some-source",
    "source_sha256": "some-source-sha",
		"cpe": "node-cpe",
		"purl": "some-purl",
		"licenses": ["MIT", "MIT-2"]
	}]`)
				}

				if req.URL.RawQuery == "name=non-existent" {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, `{"error": "error getting dependency metadata"}`)
				}

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))

		buildpackDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`
				api = "0.2"

				[buildpack]
					id = "some-buildpack"
					name = "Some Buildpack"
					version = "some-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

					[[metadata.dependencies]]
					  cpe = "node-cpe"
					  purl = "some-purl"
						id = "node"
						name = "Node Engine"
						sha256 = "some-sha"
						source = "some-source"
						source_sha256 = "some-source-sha"
						stacks = ["io.buildpacks.stacks.bionic"]
						uri = "some-dep-uri"
						version = "1.2.3"

					[[metadata.dependencies]]
					  cpe = "node-cpe"
					  purl = "some-purl"
						id = "node"
						name = "Node Engine"
						sha256 = "some-sha"
						source = "some-source"
						source_sha256 = "some-source-sha"
						stacks = ["io.buildpacks.stacks.bionic"]
						uri = "some-dep-uri"
						version = "2.1.1"

					[[metadata.dependencies]]
					  cpe = "node-cpe"
					  purl = "some-purl"
						id = "node"
						name = "Node Engine"
						sha256 = "some-sha"
						source = "some-source"
						source_sha256 = "some-source-sha"
						stacks = ["io.buildpacks.stacks.bionic"]
						uri = "some-dep-uri"
						version = "2.2.5"

				[[metadata.dependency-constraints]]
					constraint = "1.*"
					id = "node"
					patches = 1

				[[metadata.dependency-constraints]]
					constraint = "2.*"
					id = "node"
					patches = 2

				[[stacks]]
				  id = "io.buildpacks.stacks.bionic"
			`), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		server.Close()
		Expect(os.RemoveAll(buildpackDir)).To(Succeed())
	})

	it("updates the buildpack.toml dependencies", func() {
		command := exec.Command(
			path,
			"update-dependencies",
			"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
			"--api", server.URL,
		)

		buffer := gbytes.NewBuffer()
		session, err := gexec.Start(command, buffer, buffer)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

		buildpackContents, err := os.ReadFile(filepath.Join(buildpackDir, "buildpack.toml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(buildpackContents)).To(MatchTOML(`
api = "0.2"

[buildpack]
	id = "some-buildpack"
	name = "Some Buildpack"
	version = "some-buildpack-version"

[metadata]
	include-files = ["buildpack.toml"]

[[metadata.dependencies]]
	cpe = "node-cpe"
	purl = "some-purl"
	id = "node"
	licenses = ["MIT", "MIT-2"]
	name = "Node Engine"
	sha256 = "some-sha"
	source = "some-source"
	source_sha256 = "some-source-sha"
	stacks = ["io.buildpacks.stacks.bionic"]
	uri = "some-dep-uri"
	version = "1.3.5"

[[metadata.dependencies]]
	cpe = "node-cpe"
	purl = "some-purl"
	id = "node"
	licenses = ["MIT", "MIT-2"]
	name = "Node Engine"
	sha256 = "some-sha"
	source = "some-source"
	source_sha256 = "some-source-sha"
	stacks = ["io.buildpacks.stacks.bionic"]
	uri = "some-dep-uri"
	version = "2.1.9"

[[metadata.dependencies]]
	cpe = "node-cpe"
	purl = "some-purl"
	id = "node"
	licenses = ["MIT", "MIT-2"]
	name = "Node Engine"
	sha256 = "some-sha"
	source = "some-source"
	source_sha256 = "some-source-sha"
	stacks = ["io.buildpacks.stacks.bionic"]
	uri = "some-dep-uri"
	version = "2.2.5"

[[metadata.dependency-constraints]]
	constraint = "1.*"
	id = "node"
	patches = 1

[[metadata.dependency-constraints]]
	constraint = "2.*"
	id = "node"
	patches = 2

[[stacks]]
  id = "io.buildpacks.stacks.bionic"
			`))
	})

	context("the server has less patches available than requested in constraint", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`
				api = "0.2"

				[buildpack]
					id = "some-buildpack"
					name = "Some Buildpack"
					version = "some-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

				[[metadata.dependencies]]
	        cpe = "node-cpe"
					purl = "some-purl"
					id = "node"
					licenses = ["MIT", "MIT-2"]
					name = "Node Engine"
					sha256 = "some-sha"
					source = "some-source"
					source_sha256 = "some-source-sha"
					stacks = ["io.buildpacks.stacks.bionic"]
					uri = "some-dep-uri"
					version = "2.2.3"

				[[metadata.dependency-constraints]]
					constraint = "2.2.*"
					id = "node"
					patches = 3

				[[stacks]]
				  id = "io.buildpacks.stacks.bionic"
			`), 0644)

			Expect(err).ToNot(HaveOccurred())
		})
		it.After(func() {
			Expect(os.RemoveAll(buildpackDir)).To(Succeed())
		})

		it("updates the buildpack.toml dependencies with as many as are available", func() {
			command := exec.Command(
				path,
				"update-dependencies",
				"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
				"--api", server.URL,
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			buildpackContents, err := os.ReadFile(filepath.Join(buildpackDir, "buildpack.toml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buildpackContents)).To(MatchTOML(`
api = "0.2"

[buildpack]
	id = "some-buildpack"
	name = "Some Buildpack"
	version = "some-buildpack-version"

[metadata]
	include-files = ["buildpack.toml"]

[[metadata.dependencies]]
	cpe = "node-cpe"
	purl = "some-purl"
	id = "node"
	licenses = ["MIT", "MIT-2"]
	name = "Node Engine"
	sha256 = "some-sha"
	source = "some-source"
	source_sha256 = "some-source-sha"
	stacks = ["io.buildpacks.stacks.bionic"]
	uri = "some-dep-uri"
	version = "2.2.5"

[[metadata.dependency-constraints]]
	constraint = "2.2.*"
	id = "node"
	patches = 3

[[stacks]]
  id = "io.buildpacks.stacks.bionic"
			`))
		})
	})

	context("the server serves a different dependency", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`
				api = "0.2"

				[buildpack]
					id = "some-buildpack"
					name = "Some Buildpack"
					version = "some-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

				[[metadata.dependencies]]
	        cpe = "node-cpe"
					purl = "some-purl"
					id = "node"
					name = "Node Engine"
					licenses = ["MIT", "MIT-2"]
					sha256 = "some-sha"
					source = "some-source"
					source_sha256 = "some-source-sha"
					stacks = ["io.buildpacks.stacks.bionic"]
					uri = "some-dep-uri"
					version = "2.2.3"

				[[metadata.dependency-constraints]]
					constraint = "2.2.*"
					id = "node"
					patches = 3

				[[stacks]]
				  id = "io.buildpacks.stacks.bionic"
			`), 0644)

			Expect(err).ToNot(HaveOccurred())
		})
		it.After(func() {
			Expect(os.RemoveAll(buildpackDir)).To(Succeed())
		})

		it("updates the buildpack.toml dependencies with as many as are available", func() {
			command := exec.Command(
				path,
				"update-dependencies",
				"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
				"--api", server.URL,
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			buildpackContents, err := os.ReadFile(filepath.Join(buildpackDir, "buildpack.toml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buildpackContents)).To(MatchTOML(`
api = "0.2"

[buildpack]
	id = "some-buildpack"
	name = "Some Buildpack"
	version = "some-buildpack-version"

[metadata]
	include-files = ["buildpack.toml"]

[[metadata.dependencies]]
	cpe = "node-cpe"
	purl = "some-purl"
	id = "node"
	licenses = ["MIT", "MIT-2"]
	name = "Node Engine"
	sha256 = "some-sha"
	source = "some-source"
	source_sha256 = "some-source-sha"
	stacks = ["io.buildpacks.stacks.bionic"]
	uri = "some-dep-uri"
	version = "2.2.5"

[[metadata.dependency-constraints]]
	constraint = "2.2.*"
	id = "node"
	patches = 3

[[stacks]]
  id = "io.buildpacks.stacks.bionic"
			`))
		})
	})

	context("failure cases", func() {
		context("the --buildpack-file flag is missing", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-dependencies",
					"--api", server.URL,
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("Error: required flag(s) \"buildpack-file\" not set"))
			})
		})

		context("the buildpack file does not exist", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-dependencies",
					"--buildpack-file", "/no/such/file",
					"--api", server.URL,
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to parse buildpack.toml"))
			})
		})

		context("the dependencies cannot be retrieved from the server", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-dependencies",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--api", "%%%",
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to query url"))
			})
		})

		context("the server status code is not 200", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`
				api = "0.2"

				[buildpack]
					id = "some-buildpack"
					name = "Some Buildpack"
					version = "some-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

					[[metadata.dependencies]]
	          cpe = "non-existent-cpe"
						purl = "non-existent-purl"
						id = "non-existent"
					  licenses = ["MIT", "MIT-2"]
						sha256 = "some-sha"
						source = "some-source"
						source_sha256 = "some-source-sha"
						stacks = ["io.buildpacks.stacks.bionic"]
						uri = "some-dep-uri"
						version = "1.2.3"

				[[metadata.dependency-constraints]]
					constraint = "1.*"
					id = "non-existent"
					patches = 1

				[[stacks]]
				  id = "io.buildpacks.stacks.bionic"
			`), 0644)
				Expect(err).NotTo(HaveOccurred())
			})
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-dependencies",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--api", server.URL,
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring(fmt.Sprintf("failed to query url %s/v1/dependency?name=non-existent with: status code 500", server.URL)))
			})
		})
		context("when the buildpack file cannot be opened", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(buildpackDir, "buildpack.toml"), 0400)).To(Succeed())
			})
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-dependencies",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--api", server.URL,
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to open buildpack config"))
			})
		})
	})
}
