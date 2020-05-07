package main_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSummarize(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		buildpack string
		buffer    *bytes.Buffer
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack")
		Expect(err).NotTo(HaveOccurred())

		gw := gzip.NewWriter(file)
		tw := tar.NewWriter(gw)

		content := []byte(`[buildpack]
id = "some-buildpack"

[metadata.default-versions]
some-dependency = "1.2.x"
other-dependency = "2.3.x"

[[metadata.dependencies]]
	id = "some-dependency"
	stacks = ["some-stack"]
	version = "1.2.3"

[[metadata.dependencies]]
	id = "other-dependency"
	stacks = ["other-stack"]
	version = "2.3.4"

[[stacks]]
	id = "some-stack"

[[stacks]]
	id = "other-stack"
`)

		err = tw.WriteHeader(&tar.Header{
			Name: "./buildpack.toml",
			Mode: 0644,
			Size: int64(len(content)),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write(content)
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())
		Expect(gw.Close()).To(Succeed())

		buildpack = file.Name()

		Expect(file.Close()).To(Succeed())

		buffer = bytes.NewBuffer(nil)
	})

	it.After(func() {
		Expect(os.Remove(buildpack)).To(Succeed())
	})

	it("prints out the summary of a buildpack tarball", func() {
		command := exec.Command(
			path, "summarize",
			"--buildpack", buildpack,
			"--format", "markdown",
		)
		session, err := gexec.Start(command, buffer, buffer)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0), func() string { return buffer.String() })

		Expect(session.Out).To(gbytes.Say("Dependencies:"))
		Expect(session.Out).To(gbytes.Say("| name | version | stacks |"))
		Expect(session.Out).To(gbytes.Say("|-|-|-|"))
		Expect(session.Out).To(gbytes.Say("| other-dependency | 2.3.4 | other-stack |"))
		Expect(session.Out).To(gbytes.Say("| some-dependency | 1.2.3 | some-stack |"))

		Expect(session.Out).To(gbytes.Say("Default dependencies:"))
		Expect(session.Out).To(gbytes.Say("| name | version |"))
		Expect(session.Out).To(gbytes.Say("|-|-|"))
		Expect(session.Out).To(gbytes.Say("| other-dependency | 2.3.x |"))
		Expect(session.Out).To(gbytes.Say("| some-dependency | 1.2.x |"))

		Expect(session.Out).To(gbytes.Say("Supported stacks:"))
		Expect(session.Out).To(gbytes.Say("| name |"))
		Expect(session.Out).To(gbytes.Say("|-|"))
		Expect(session.Out).To(gbytes.Say("| other-stack |"))
		Expect(session.Out).To(gbytes.Say("| some-stack |"))
	})
}
