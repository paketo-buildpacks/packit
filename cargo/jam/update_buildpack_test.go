package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testUpdateBuildpack(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		server       *httptest.Server
		buildpackDir string
	)

	it.Before(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)

				return
			}

			switch req.URL.Path {
			case "/v2/":
				w.WriteHeader(http.StatusOK)

			case "/v2/some-repository/some-buildpack-id/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.20.1",
								"0.20.12",
								"latest"
							]
					}`)

			case "/v2/some-repository/other-buildpack-id/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.1.0",
								"0.20.2",
								"0.20.22"
							]
					}`)

			case "/v2/some-repository/last-buildpack-id/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.1.0",
								"0.2.0",
								"latest"
							]
					}`)

			case "/v2/some-repository/error-buildpack-id/tags/list":
				w.WriteHeader(http.StatusTeapot)

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))

		var err error
		buildpackDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`
				api = "0.2"

				[buildpack]
					id = "some-composite-buildpack"
					name = "Some Composite Buildpack"
					version = "some-composite-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

				[[order]]
					[[order.group]]
						id = "some-repository/some-buildpack-id"
						version = "0.20.1"

					[[order.group]]
						id = "some-repository/last-buildpack-id"
						version = "0.2.0"

				[[order]]
					[[order.group]]
						id = "some-repository/other-buildpack-id"
						version = "0.1.0"
						optional = true

					[[order.group]]
						id = "some-repository/some-buildpack-id"
						version = "0.20.1"
			`), 0600)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(buildpackDir, "package.toml"), bytes.ReplaceAll([]byte(`
				[buildpack]
				uri = "build/buildpack.tgz"

				[[dependencies]]
				uri = "docker://REGISTRY-URI/some-repository/last-buildpack-id:0.2.0"

				[[dependencies]]
				uri = "docker://REGISTRY-URI/some-repository/some-buildpack-id:0.20.1"

				[[dependencies]]
				uri = "docker://REGISTRY-URI/some-repository/other-buildpack-id:0.1.0"
			`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		server.Close()
	})

	it("updates the buildpack.toml and package.toml files", func() {
		command := exec.Command(
			path,
			"update-buildpack",
			"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
			"--package-file", filepath.Join(buildpackDir, "package.toml"),
		)

		buffer := gbytes.NewBuffer()
		session, err := gexec.Start(command, buffer, buffer)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

		buildpackContents, err := ioutil.ReadFile(filepath.Join(buildpackDir, "buildpack.toml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(buildpackContents)).To(MatchTOML(`
				api = "0.2"

				[buildpack]
					id = "some-composite-buildpack"
					name = "Some Composite Buildpack"
					version = "some-composite-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

				[[order]]
					[[order.group]]
						id = "some-repository/some-buildpack-id"
						version = "0.20.12"

					[[order.group]]
						id = "some-repository/last-buildpack-id"
						version = "0.2.0"

				[[order]]
					[[order.group]]
						id = "some-repository/other-buildpack-id"
						version = "0.20.22"
						optional = true

					[[order.group]]
						id = "some-repository/some-buildpack-id"
						version = "0.20.12"
			`))

		packageContents, err := ioutil.ReadFile(filepath.Join(buildpackDir, "package.toml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(packageContents)).To(MatchTOML(strings.ReplaceAll(`
				[buildpack]
				uri = "build/buildpack.tgz"

				[[dependencies]]
				uri = "docker://REGISTRY-URI/some-repository/last-buildpack-id:0.2.0"

				[[dependencies]]
				uri = "docker://REGISTRY-URI/some-repository/some-buildpack-id:0.20.12"

				[[dependencies]]
				uri = "docker://REGISTRY-URI/some-repository/other-buildpack-id:0.20.22"
			`, "REGISTRY-URI", strings.TrimPrefix(server.URL, "http://"))))
	})

	context("failure cases", func() {
		context("when the --buildpack-file flag is missing", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--package-file", filepath.Join(buildpackDir, "package.toml"),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(Equal("failed to execute: --buildpack-file is a required flag"))
			})
		})

		context("when the --package-file flag is missing", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(Equal("failed to execute: --package-file is a required flag"))
			})
		})

		context("when the buildpack file does not exist", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--buildpack-file", "/no/such/file",
					"--package-file", filepath.Join(buildpackDir, "package.toml"),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(Equal("failed to execute: failed to open buildpack config file: open /no/such/file: no such file or directory"))
			})
		})

		context("when the package file does not exist", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--package-file", "/no/such/file",
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(Equal("failed to execute: failed to open package config file: open /no/such/file: no such file or directory"))
			})
		})

		context("when the latest image reference cannot be found", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`
						api = "0.2"

						[buildpack]
							id = "some-composite-buildpack"
							name = "Some Composite Buildpack"
							version = "some-composite-buildpack-version"

						[metadata]
							include-files = ["buildpack.toml"]

						[[order]]
							[[order.group]]
								id = "some-repository/error-buildpack-id"
								version = "0.20.1"
					`), 0600)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(buildpackDir, "package.toml"), bytes.ReplaceAll([]byte(`
						[buildpack]
						uri = "build/buildpack.tgz"

						[[dependencies]]
						uri = "docker://REGISTRY-URI/some-repository/error-buildpack-id:0.20.1"
					`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--package-file", filepath.Join(buildpackDir, "package.toml"),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to list tags"))
			})
		})

		context("when the buildpack file cannot be overwritten", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(buildpackDir, "buildpack.toml"), 0400)).To(Succeed())
			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--package-file", filepath.Join(buildpackDir, "package.toml"),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to open buildpack config"))
			})
		})

		context("when the package file cannot be overwritten", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(buildpackDir, "package.toml"), 0400)).To(Succeed())
			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-buildpack",
					"--buildpack-file", filepath.Join(buildpackDir, "buildpack.toml"),
					"--package-file", filepath.Join(buildpackDir, "package.toml"),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to open package config"))
			})
		})
	})
}
