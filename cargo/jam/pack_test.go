package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testPack(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		buffer       *Buffer
		tmpDir       string
		buildpackDir string
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "output")
		Expect(err).NotTo(HaveOccurred())

		buildpackDir, err = os.MkdirTemp("", "buildpack")
		Expect(err).NotTo(HaveOccurred())

		buffer = &Buffer{}
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(buildpackDir)).To(Succeed())
	})

	context("when packaging a language family buildpack", func() {

		it.Before(func() {
			err := cargo.NewDirectoryDuplicator().Duplicate(filepath.Join("testdata", "example-language-family-cnb"), buildpackDir)
			Expect(err).NotTo(HaveOccurred())
		})

		it("creates a language family archive", func() {
			command := exec.Command(
				path, "pack",
				"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
				"--output", filepath.Join(tmpDir, "output.tgz"),
				"--version", "some-version",
			)
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5s").Should(gexec.Exit(0), func() string { return buffer.String() })

			Expect(session.Out).To(gbytes.Say("Packing some-buildpack-name some-version..."))
			Expect(session.Out).To(gbytes.Say(fmt.Sprintf("  Building tarball: %s", filepath.Join(tmpDir, "output.tgz"))))
			Expect(session.Out).To(gbytes.Say("    buildpack.toml"))

			file, err := os.Open(filepath.Join(tmpDir, "output.tgz"))
			Expect(err).NotTo(HaveOccurred())

			contents, _, err := ExtractFile(file, "buildpack.toml")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchTOML(`api = "0.2"

[buildpack]
  id = "some-buildpack-id"
  name = "some-buildpack-name"
  version = "some-version"

[metadata]
  include-files = ["buildpack.toml"]

  [[metadata.dependencies]]
    deprecation_date = "2019-04-01T00:00:00Z"
    id = "some-dependency"
    name = "Some Dependency"
    sha256 = "shasum"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    uri = "http://some-url"
    version = "1.2.3"

  [[metadata.dependencies]]
		deprecation_date = "2022-04-01T00:00:00Z"
    id = "other-dependency"
    name = "Other Dependency"
    sha256 = "shasum"
    stacks = ["org.cloudfoundry.stacks.tiny"]
    uri = "http://other-url"
    version = "4.5.6"

[[order]]
  [[order.group]]
    id = "some-dependency"
    version = "1.2.3"

  [[order.group]]
    id = "other-dependency"
    version = "4.5.6"`))
		})
	})

	context("when packaging an implementation buildpack", func() {
		it.Before(func() {
			err := cargo.NewDirectoryDuplicator().Duplicate(filepath.Join("testdata", "example-cnb"), buildpackDir)
			Expect(err).NotTo(HaveOccurred())
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
			Eventually(session, "5s").Should(gexec.Exit(0), func() string { return buffer.String() })

			Expect(session.Out).To(gbytes.Say("Packing some-buildpack-name some-version..."))
			Expect(session.Out).To(gbytes.Say("  Executing pre-packaging script: ./scripts/build.sh"))
			Expect(session.Out).To(gbytes.Say("    hello from the pre-packaging script"))
			Expect(session.Out).To(gbytes.Say(fmt.Sprintf("  Building tarball: %s", filepath.Join(tmpDir, "output.tgz"))))
			Expect(session.Out).To(gbytes.Say("    bin/build"))
			Expect(session.Out).To(gbytes.Say("    bin/detect"))
			Expect(session.Out).To(gbytes.Say("    bin/link"))
			Expect(session.Out).To(gbytes.Say("    buildpack.toml"))
			Expect(session.Out).To(gbytes.Say("    generated-file"))

			file, err := os.Open(filepath.Join(tmpDir, "output.tgz"))
			Expect(err).NotTo(HaveOccurred())

			u, err := user.Current()
			Expect(err).NotTo(HaveOccurred())
			userName := u.Username

			group, err := user.LookupGroupId(u.Gid)
			Expect(err).NotTo(HaveOccurred())
			groupName := group.Name

			contents, hdr, err := ExtractFile(file, "buildpack.toml")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchTOML(`api = "0.2"

[buildpack]
  id = "some-buildpack-id"
  name = "some-buildpack-name"
  version = "some-version"
  homepage = "some-homepage-link"

[metadata]
  include-files = ["bin/build", "bin/detect", "bin/link", "buildpack.toml", "generated-file"]
  pre-package = "./scripts/build.sh"
  [metadata.default-versions]
    some-dependency = "some-default-version"

  [[metadata.dependencies]]
    deprecation_date = "2019-04-01T00:00:00Z"
    id = "some-dependency"
    name = "Some Dependency"
    sha256 = "shasum"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    uri = "http://some-url"
    version = "1.2.3"

  [[metadata.dependencies]]
    deprecation_date = "2022-04-01T00:00:00Z"
    id = "other-dependency"
    name = "Other Dependency"
    sha256 = "shasum"
    stacks = ["org.cloudfoundry.stacks.tiny"]
    uri = "http://other-url"
    version = "4.5.6"

[[stacks]]
  id = "some-stack-id"
  mixins = ["some-mixin-id"]`))
			Expect(hdr.Mode).To(Equal(int64(0644)))

			contents, hdr, err = ExtractFile(file, "bin/build")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("build-contents"))
			Expect(hdr.Mode).To(Equal(int64(0755)))
			Expect(hdr.Uname).To(Equal(userName))
			Expect(hdr.Gname).To(Equal(groupName))

			contents, hdr, err = ExtractFile(file, "bin/detect")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("detect-contents"))
			Expect(hdr.Mode).To(Equal(int64(0755)))
			Expect(hdr.Uname).To(Equal(userName))
			Expect(hdr.Gname).To(Equal(groupName))

			_, hdr, err = ExtractFile(file, "bin/link")
			Expect(err).NotTo(HaveOccurred())
			Expect(hdr.Linkname).To(Equal("build"))
			Expect(hdr.Uname).To(Equal(userName))
			Expect(hdr.Gname).To(Equal(groupName))

			contents, hdr, err = ExtractFile(file, "generated-file")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("hello\n"))
			Expect(hdr.Mode).To(Equal(int64(0644)))
			Expect(hdr.Uname).To(Equal(userName))
			Expect(hdr.Gname).To(Equal(groupName))

			Expect(filepath.Join(buildpackDir, "generated-file")).NotTo(BeARegularFile())
		})

		context("when the buildpack is built to run offline", func() {
			var server *httptest.Server
			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.URL.Path != "/some-dependency.tgz" {
						http.NotFound(w, req)
					}

					fmt.Fprint(w, "dependency-contents")
				}))

				config, err := cargo.NewBuildpackParser().Parse(filepath.Join(buildpackDir, "buildpack.toml"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Metadata.Dependencies).To(HaveLen(2))

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
					"--stack", "io.buildpacks.stacks.bionic",
				)
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), func() string { return buffer.String() })

				Expect(session.Out).To(gbytes.Say("Packing some-buildpack-name some-version..."))
				Expect(session.Out).To(gbytes.Say("  Executing pre-packaging script: ./scripts/build.sh"))
				Expect(session.Out).To(gbytes.Say("    hello from the pre-packaging script"))
				Expect(session.Out).To(gbytes.Say("  Downloading dependencies..."))
				Expect(session.Out).To(gbytes.Say(`    some-dependency \(1.2.3\) \[io.buildpacks.stacks.bionic, org.cloudfoundry.stacks.tiny\]`))
				Expect(session.Out).To(gbytes.Say("      ↳  dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"))
				Expect(session.Out).To(gbytes.Say(fmt.Sprintf("  Building tarball: %s", filepath.Join(tmpDir, "output.tgz"))))
				Expect(session.Out).To(gbytes.Say("    bin"))
				Expect(session.Out).To(gbytes.Say("    bin/build"))
				Expect(session.Out).To(gbytes.Say("    bin/detect"))
				Expect(session.Out).To(gbytes.Say("    buildpack.toml"))
				Expect(session.Out).To(gbytes.Say("    dependencies"))
				Expect(session.Out).To(gbytes.Say("    dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"))
				Expect(session.Out).To(gbytes.Say("    generated-file"))

				Expect(string(session.Out.Contents())).NotTo(ContainSubstring("other-dependency"))

				file, err := os.Open(filepath.Join(tmpDir, "output.tgz"))
				Expect(err).NotTo(HaveOccurred())

				contents, hdr, err := ExtractFile(file, "buildpack.toml")
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(MatchTOML(`api = "0.2"

[buildpack]
  id = "some-buildpack-id"
  name = "some-buildpack-name"
  version = "some-version"
  homepage = "some-homepage-link"

[metadata]
  include-files = ["bin/build", "bin/detect", "bin/link", "buildpack.toml", "generated-file", "dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"]
  pre-package = "./scripts/build.sh"
  [metadata.default-versions]
    some-dependency = "some-default-version"

  [[metadata.dependencies]]
    deprecation_date = "2019-04-01T00:00:00Z"
    id = "some-dependency"
    name = "Some Dependency"
    sha256 = "f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    uri = "file:///dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a"
    version = "1.2.3"

[[stacks]]
  id = "some-stack-id"
  mixins = ["some-mixin-id"]`))
				Expect(hdr.Mode).To(Equal(int64(0644)))

				contents, hdr, err = ExtractFile(file, "bin/build")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("build-contents"))
				Expect(hdr.Mode).To(Equal(int64(0755)))

				contents, hdr, err = ExtractFile(file, "bin/detect")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("detect-contents"))
				Expect(hdr.Mode).To(Equal(int64(0755)))

				contents, hdr, err = ExtractFile(file, "generated-file")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("hello\n"))
				Expect(hdr.Mode).To(Equal(int64(0644)))

				contents, hdr, err = ExtractFile(file, "dependencies/f058c8bf6b65b829e200ef5c2d22fde0ee65b96c1fbd1b88869be133aafab64a")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("dependency-contents"))
				Expect(hdr.Mode).To(Equal(int64(0644)))
			})
		})
	})

	context("failure cases", func() {
		context("when the all the required flags are not set", func() {
			it("prints an error message", func() {
				command := exec.Command(path, "pack")
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err.Contents()).To(ContainSubstring("Error: required flag(s) \"buildpack\", \"output\", \"version\" not set"))
			})
		})

		context("when the required buildpack flag is not set", func() {
			it("prints an error message", func() {
				command := exec.Command(
					path, "pack",
					"--output", filepath.Join(tmpDir, "output.tgz"),
					"--version", "some-version",
					"--offline",
					"--stack", "io.buildpacks.stacks.bionic",
				)
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err.Contents()).To(ContainSubstring("Error: required flag(s) \"buildpack\" not set"))
			})
		})

		context("when the required output flag is not set", func() {
			it("prints an error message", func() {
				command := exec.Command(
					path, "pack",
					"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
					"--version", "some-version",
					"--offline",
					"--stack", "io.buildpacks.stacks.bionic",
				)
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err.Contents()).To(ContainSubstring("Error: required flag(s) \"output\" not set"))
			})
		})

		context("when the required version flag is not set", func() {
			it("prints an error message", func() {
				command := exec.Command(
					path, "pack",
					"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
					"--output", filepath.Join(tmpDir, "output.tgz"),
					"--offline",
					"--stack", "io.buildpacks.stacks.bionic",
				)
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err.Contents()).To(ContainSubstring("Error: required flag(s) \"version\" not set"))
			})
		})
	})
}
