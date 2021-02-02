package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func testUpdateBuilder(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		server     *httptest.Server
		builderDir string
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

			case "/v2/some-repository/error-buildpack-id/tags/list":
				w.WriteHeader(http.StatusTeapot)

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))

		var err error
		builderDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(builderDir, "builder.toml"), bytes.ReplaceAll([]byte(`
description = "Some description"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/some-repository/some-buildpack-id:0.0.10"
  version = "0.0.10"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/some-repository/other-buildpack-id:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.10.2"

[[order]]

  [[order.group]]
    id = "some-repository/other-buildpack-id"
    version = "0.20.22"

[[order]]

  [[order.group]]
    id = "some-repository/some-buildpack-id"
    version = "0.0.10"

[stack]
  id = "io.paketo.stacks.some-stack"
  build-image = "REGISTRY-URI/somerepository/build:1.2.3-some-cnb"
  run-image = "REGISTRY-URI.io/somerepository/run:some-cnb"
  run-image-mirrors = ["REGISTRY-URI/some-repository/run:some-cnb"]
			`), []byte(`REGISTRY-URI`), []byte(strings.TrimPrefix(server.URL, "http://"))), 0600)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		server.Close()
	})

	it("updates the builder files", func() {
		command := exec.Command(
			path,
			"update-builder",
			"--builder-file", filepath.Join(builderDir, "builder.toml"),
			"--lifecycle-uri", fmt.Sprintf("%s/some-repository/lifecycle", server.URL),
		)

		buffer := gbytes.NewBuffer()
		session, err := gexec.Start(command, buffer, buffer)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

		builderContents, err := ioutil.ReadFile(filepath.Join(builderDir, "builder.toml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(builderContents)).To(MatchTOML(strings.ReplaceAll(`
description = "Some description"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/some-repository/some-buildpack-id:0.20.12"
  version = "0.20.12"

[[buildpacks]]
	uri = "docker://REGISTRY-URI/some-repository/other-buildpack-id:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.21.1"

[[order]]

  [[order.group]]
    id = "some-repository/other-buildpack-id"
    version = "0.20.22"

[[order]]

  [[order.group]]
    id = "some-repository/some-buildpack-id"
    version = "0.20.12"

[stack]
  id = "io.paketo.stacks.some-stack"
  build-image = "REGISTRY-URI/somerepository/build:1.2.3-some-cnb"
  run-image = "REGISTRY-URI.io/somerepository/run:some-cnb"
  run-image-mirrors = ["REGISTRY-URI/some-repository/run:some-cnb"]
			`, "REGISTRY-URI", strings.TrimPrefix(server.URL, "http://"))))
	})
}
