package main_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/packit/matchers"
	. "github.com/onsi/gomega"
)

func testPack(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		buffer *bytes.Buffer
		tmpDir string
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)

		var err error
		tmpDir, err = ioutil.TempDir("", "output")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(filepath.Join("testdata", "example-cnb", "generated-file"))).To(Succeed())
	})

	it("creates a packaged buildpack", func() {
		command := exec.Command(
			path, "pack",
			"--buildpack", filepath.Join("testdata", "example-cnb", "buildpack.toml"),
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
pre_package = "./scripts/build.sh"`))
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
