package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func mustConfigName(t *testing.T, img v1.Image) v1.Hash {
	h, err := img.ConfigName()
	if err != nil {
		t.Fatalf("ConfigName() = %v", err)
	}
	return h
}

func mustRawManifest(t *testing.T, img remote.Taggable) []byte {
	m, err := img.RawManifest()
	if err != nil {
		t.Fatalf("RawManifest() = %v", err)
	}
	return m
}

func mustRawConfigFile(t *testing.T, img v1.Image) []byte {
	c, err := img.RawConfigFile()
	if err != nil {
		t.Fatalf("RawConfigFile() = %v", err)
	}
	return c
}

func testUpdateBuilder(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		server     *httptest.Server
		builderDir string
	)

	it.Before(func() {
		goRef, err := name.ParseReference("gcr.io/paketo-buildpacks/go")
		Expect(err).ToNot(HaveOccurred())
		goImg, err := remote.Image(goRef)
		Expect(err).ToNot(HaveOccurred())

		nodeRef, err := name.ParseReference("paketobuildpacks/nodejs")
		Expect(err).ToNot(HaveOccurred())
		nodeImg, err := remote.Image(nodeRef)
		Expect(err).ToNot(HaveOccurred())

		goManifestPath := "/v2/paketo-buildpacks/go/manifests/0.0.10"
		goConfigPath := fmt.Sprintf("/v2/paketo-buildpacks/go/blobs/%s", mustConfigName(t, goImg))
		goManifestReqCount := 0
		nodeManifestPath := "/v2/paketobuildpacks/nodejs/manifests/0.20.22"
		nodeConfigPath := fmt.Sprintf("/v2/paketobuildpacks/nodejs/blobs/%s", mustConfigName(t, nodeImg))
		nodeManifestReqCount := 0
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)

				return
			}

			switch req.URL.Path {
			case "/v2/":
				w.WriteHeader(http.StatusOK)

			case "/v2/paketo-buildpacks/go/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.20.1",
								"0.20.12",
								"latest"
							]
					}`)

			case "/v2/paketobuildpacks/nodejs/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.1.0",
								"0.20.2",
								"0.20.22"
							]
					}`)

			case "/v2/some-repository/lifecycle/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.20.1",
								"0.21.1",
								"latest"
							]
					}`)

			case "/v2/somerepository/build/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10-some-cnb",
								"0.20.1",
								"0.20.12-some-cnb",
								"0.20.12-other-cnb",
								"latest"
							]
					}`)

			case goConfigPath:
				if req.Method != http.MethodGet {
					t.Errorf("Method; got %v, want %v", req.Method, http.MethodGet)
				}
				_, _ = w.Write(mustRawConfigFile(t, goImg))

			case goManifestPath:
				goManifestReqCount++
				if req.Method != http.MethodGet {
					t.Errorf("Method; got %v, want %v", req.Method, http.MethodGet)
				}
				_, _ = w.Write(mustRawManifest(t, goImg))

			case nodeConfigPath:
				if req.Method != http.MethodGet {
					t.Errorf("Method; got %v, want %v", req.Method, http.MethodGet)
				}
				_, _ = w.Write(mustRawConfigFile(t, nodeImg))

			case nodeManifestPath:
				nodeManifestReqCount++
				if req.Method != http.MethodGet {
					t.Errorf("Method; got %v, want %v", req.Method, http.MethodGet)
				}
				_, _ = w.Write(mustRawManifest(t, nodeImg))

			case "/v2/some-repository/error-buildpack-id/tags/list":
				w.WriteHeader(http.StatusTeapot)

			case "/v2/some-repository/error-lifecycle/tags/list":
				w.WriteHeader(http.StatusTeapot)

			case "/v2/somerepository/error-build/tags/list":
				w.WriteHeader(http.StatusTeapot)

			case "/v2/some-repository/nonexistent-labels-id/tags/list":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{
						  "tags": [
								"0.1.0",
								"0.2.0",
								"latest"
							]
					}`)

			case "/v2/some-repository/nonexistent-labels-id/manifests/0.2.0":
				w.WriteHeader(http.StatusBadRequest)

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))

		builderDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(builderDir, "builder.toml"), bytes.ReplaceAll([]byte(`
description = "Some description"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/paketo-buildpacks/go:0.0.10"
  version = "0.0.10"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/paketobuildpacks/nodejs:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.10.2"

[[order]]

  [[order.group]]
    id = "paketo-buildpacks/nodejs"

[[order]]

  [[order.group]]
		id = "paketo-buildpacks/go"
    version = "0.0.10"
		optional = true

[stack]
  id = "io.paketo.stacks.some-stack"
  build-image = "REGISTRY-URI/somerepository/build:0.0.10-some-cnb"
  run-image = "REGISTRY-URI/somerepository/run:some-cnb"
  run-image-mirrors = ["REGISTRY-URI/some-repository/run:some-cnb"]
			`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		server.Close()
		Expect(os.RemoveAll(builderDir)).To(Succeed())
	})

	it("updates the builder files", func() {
		command := exec.Command(
			path,
			"update-builder",
			"--builder-file", filepath.Join(builderDir, "builder.toml"),
			"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
		)

		buffer := gbytes.NewBuffer()
		session, err := gexec.Start(command, buffer, buffer)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

		builderContents, err := os.ReadFile(filepath.Join(builderDir, "builder.toml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(builderContents)).To(MatchTOML(strings.ReplaceAll(`
description = "Some description"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/paketo-buildpacks/go:0.20.12"
  version = "0.20.12"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/paketobuildpacks/nodejs:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.21.1"

[[order]]

  [[order.group]]
    id = "paketo-buildpacks/nodejs"

[[order]]

  [[order.group]]
    id = "paketo-buildpacks/go"
		version = "0.20.12"
		optional = true

[stack]
  id = "io.paketo.stacks.some-stack"
  build-image = "REGISTRY-URI/somerepository/build:0.20.12-some-cnb"
  run-image = "REGISTRY-URI/somerepository/run:some-cnb"
  run-image-mirrors = ["REGISTRY-URI/some-repository/run:some-cnb"]
			`, "REGISTRY-URI", strings.TrimPrefix(server.URL, "http://"))))
	})

	context("failure cases", func() {
		context("when the --builder-file flag is missing", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("Error: required flag(s) \"builder-file\" not set"))
			})
		})

		context("when the builder file cannot be found", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--builder-file", "/no/such/file",
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to execute: failed to open builder config file: open /no/such/file: no such file or directory"))
			})
		})

		context("when the latest buildpack image cannot be found", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(builderDir, "builder.toml"), bytes.ReplaceAll([]byte(`
[[buildpacks]]
	uri = "docker://REGISTRY-URI/some-repository/error-buildpack-id:0.0.10"
  version = "0.0.10"
			`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--builder-file", filepath.Join(builderDir, "builder.toml"),
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to list tags"))
			})
		})

		context("when the latest lifecycle image cannot be found", func() {
			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--builder-file", filepath.Join(builderDir, "builder.toml"),
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/error-lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to list tags"))
			})
		})

		context("when the buildpackage ID cannot be retrieved", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(builderDir, "builder.toml"), bytes.ReplaceAll([]byte(`
description = "Some description"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/some-repository/nonexistent-labels-id:0.2.0"
  version = "0.2.0"

[lifecycle]
  version = "0.10.2"

[[order]]

  [[order.group]]
		id = "some-repository/nonexistent-labels-id"
  	version = "0.2.0"

[stack]
  id = "io.paketo.stacks.some-stack"
	build-image = "REGISTRY-URI/somerepository/error-build:0.0.10-some-cnb"
  run-image = "REGISTRY-URI/somerepository/run:some-cnb"
  run-image-mirrors = ["REGISTRY-URI/some-repository/run:some-cnb"]
			`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
				Expect(err).NotTo(HaveOccurred())

			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--builder-file", filepath.Join(builderDir, "builder.toml"),
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(MatchRegexp(`failed to get buildpackage ID for \d+\.\d+\.\d+\.\d+\:\d+\/some\-repository\/nonexistent\-labels\-id\:0\.2\.0\:`))
				Expect(string(buffer.Contents())).To(ContainSubstring("unexpected status code 400 Bad Request"))
			})
		})

		context("when the latest build image cannot be found", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(builderDir, "builder.toml"), bytes.ReplaceAll([]byte(`
description = "Some description"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/paketo-buildpacks/go:0.0.10"
  version = "0.0.10"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/paketobuildpacks/nodejs:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.10.2"

[[order]]

  [[order.group]]
    id = "paketo-buildpacks/nodejs"
    version = "0.20.22"

[[order]]

  [[order.group]]
    id = "paketo-buildpacks/go"
    version = "0.0.10"

[stack]
  id = "io.paketo.stacks.some-stack"
	build-image = "REGISTRY-URI/somerepository/error-build:0.0.10-some-cnb"
  run-image = "REGISTRY-URI/somerepository/run:some-cnb"
  run-image-mirrors = ["REGISTRY-URI/some-repository/run:some-cnb"]
			`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--builder-file", filepath.Join(builderDir, "builder.toml"),
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to list tags"))
			})
		})

		context("when the builder file cannot be overwritten", func() {
			it.Before(func() {
				err := os.Chmod(filepath.Join(builderDir, "builder.toml"), 0400)
				Expect(err).NotTo(HaveOccurred())
			})

			it("prints an error and exits non-zero", func() {
				command := exec.Command(
					path,
					"update-builder",
					"--builder-file", filepath.Join(builderDir, "builder.toml"),
					"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", strings.TrimPrefix(server.URL, "http://")),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(string(buffer.Contents())).To(ContainSubstring("failed to open builder config"))
			})
		})
	})
}
