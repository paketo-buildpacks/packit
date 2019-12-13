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
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	"github.com/cloudfoundry/packit/cargo"
	. "github.com/cloudfoundry/packit/matchers"
	. "github.com/onsi/gomega"
)

func testPack(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		buffer       *bytes.Buffer
		tmpDir       string
		buildpackDir string
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)

		var err error
		tmpDir, err = ioutil.TempDir("", "output")
		Expect(err).NotTo(HaveOccurred())

		buildpackDir, err = ioutil.TempDir("", "buildpack")
		Expect(err).NotTo(HaveOccurred())

		err = cargo.NewDirectoryDuplicator().Duplicate(filepath.Join("testdata", "example-cnb"), buildpackDir)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(buildpackDir)).To(Succeed())
	})

	it("creates a packaged buildpack", func() {
		command := exec.Command(
			path, "pack",
			"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
			"--output", filepath.Join(tmpDir, "output.tgz"),
			"--version", "some-version",
		)
		session, err := gexec.Start(command, buffer, buffer)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0), func() string { return buffer.String() })

		Expect(session.Out).To(gbytes.Say("Packing some-buildpack-name some-version..."))

		file, err := os.Open(filepath.Join(tmpDir, "output.tgz"))
		Expect(err).ToNot(HaveOccurred())

		contents, hdr, err := ExtractFile(file, "buildpack.toml")
		Expect(err).ToNot(HaveOccurred())
		Expect(contents).To(MatchTOML(`api = "0.2"

[buildpack]
  id = "some-buildpack-id"
  name = "some-buildpack-name"
  version = "some-version"

[metadata]
  include_files = ["bin/build", "bin/detect", "buildpack.toml", "generated-file"]
  pre_package = "./scripts/build.sh"
  [metadata.default-versions]
    some-dependency = "some-default-version"

  [[metadata.dependencies]]
    id = "some-dependency"
    name = "Some Dependency"
    sha256 = "shasum"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    uri = "http://some-url"
    version = "1.2.3"

[[stacks]]
  id = "some-stack-id"`))
		Expect(hdr.Mode).To(Equal(int64(0644)))

		contents, hdr, err = ExtractFile(file, "bin/build")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(contents)).To(Equal("build-contents"))
		Expect(hdr.Mode).To(Equal(int64(0755)))

		contents, hdr, err = ExtractFile(file, "bin/detect")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(contents)).To(Equal("detect-contents"))
		Expect(hdr.Mode).To(Equal(int64(0755)))

		contents, hdr, err = ExtractFile(file, "generated-file")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(contents)).To(Equal("hello\n"))
		Expect(hdr.Mode).To(Equal(int64(0644)))

		Expect(filepath.Join(buildpackDir, "generated-file")).NotTo(BeARegularFile())
	})

	context("when the buildpack is built to run offline", func() {
		var server *httptest.Server
		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/some-dependency.tgz" {
					http.NotFound(w, req)
				}
				w.Write([]byte("dependency-contents"))
			}))

			config, err := cargo.NewBuildpackParser().Parse(filepath.Join(buildpackDir, "buildpack.toml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Metadata.Dependencies).To(HaveLen(1))

			config.Metadata.Dependencies[0].URI = fmt.Sprintf("%s/some-dependency.tgz", server.URL)
			config.Metadata.Dependencies[0].SHA256 = "f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"

			bpTomlWriter, err := os.Create(filepath.Join(buildpackDir, "buildpack.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(cargo.EncodeConfig(bpTomlWriter, config)).To(Succeed())
		})

		it.After(func() {
			server.Close()
		})

		it("creates an offline packaged buildpack", func() {
			command := exec.Command(
				path, "pack",
				"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
				"--output", filepath.Join(tmpDir, "output.tgz"),
				"--version", "some-version",
				"--offline",
			)
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), func() string { return buffer.String() })

			Expect(session.Out).To(gbytes.Say("Packing some-buildpack-name some-version..."))

			file, err := os.Open(filepath.Join(tmpDir, "output.tgz"))
			Expect(err).ToNot(HaveOccurred())

			contents, hdr, err := ExtractFile(file, "buildpack.toml")
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchTOML(`api = "0.2"

[buildpack]
  id = "some-buildpack-id"
  name = "some-buildpack-name"
  version = "some-version"

[metadata]
  include_files = ["bin/build", "bin/detect", "buildpack.toml", "generated-file", "dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"]
  pre_package = "./scripts/build.sh"
  [metadata.default-versions]
    some-dependency = "some-default-version"

  [[metadata.dependencies]]
    id = "some-dependency"
    name = "Some Dependency"
    sha256 = "f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    uri = "file:///dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"
    version = "1.2.3"

[[stacks]]
  id = "some-stack-id"`))
			Expect(hdr.Mode).To(Equal(int64(0644)))

			contents, hdr, err = ExtractFile(file, "bin/build")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("build-contents"))
			Expect(hdr.Mode).To(Equal(int64(0755)))

			contents, hdr, err = ExtractFile(file, "bin/detect")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("detect-contents"))
			Expect(hdr.Mode).To(Equal(int64(0755)))

			contents, hdr, err = ExtractFile(file, "generated-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("hello\n"))
			Expect(hdr.Mode).To(Equal(int64(0644)))

			contents, hdr, err = ExtractFile(file, "dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("dependency-contents"))
			Expect(hdr.Mode).To(Equal(int64(0644)))
		})
	})

	context("failure cases", func() {
		context("when the --buildpack flag is empty", func() {
			it("prints an error message", func() {
				command := exec.Command(path, "pack")
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err).To(gbytes.Say("missing required flag --buildpack"))
			})
		})
	})
}
