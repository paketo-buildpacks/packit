package sbom_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testAutobom(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect

		path        string
		buffer      *bytes.Buffer
		clock       chronos.Clock
		logger      scribe.Emitter
		sbomFormats []string
	)

	it.Before(func() {
		var err error
		path, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		err = fs.Copy(filepath.Join("testdata", "package-lock.json"), filepath.Join(path, "package-lock.json"))
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)

		clock = chronos.NewClock(func() time.Time {
			return time.UnixMicro(998877)
		})

		sbomFormats = []string{"application/vnd.cyclonedx+json", "application/spdx+json", "application/vnd.syft+json"}
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("returns the SBOM for the path", func() {
		sbomFormatter, err := sbom.Autobom(path, sbomFormats, clock, logger)
		Expect(err).ToNot(HaveOccurred())

		// NOTE sbom generation is NOT REPRODUCIBLE
		// Use some checks here to gain confidence that the Autobom func did what it was supposed to do
		Expect(len(sbomFormatter.Formats())).To(Equal(len(sbomFormats)))

		var extensions []string
		for _, item := range sbomFormatter.Formats() {
			extensions = append(extensions, item.Extension)
		}

		Expect(extensions).To(Equal([]string{"cdx.json", "spdx.json", "syft.json"}))
	})

	it("logs information about the formats", func() {
		_, err := sbom.Autobom(path, sbomFormats, clock, logger)
		Expect(err).ToNot(HaveOccurred())

		Expect(buffer.String()).To(Equal(fmt.Sprintf(`  Generating SBOM for %s
      Completed in 0s

`, path)))
	})

	context("with debug logging", func() {
		it.Before(func() {
			logger = scribe.NewEmitter(buffer).WithLevel("DEBUG")
		})

		it("logs information about the formats", func() {
			_, err := sbom.Autobom(path, sbomFormats, clock, logger)
			Expect(err).ToNot(HaveOccurred())

			Expect(buffer.String()).To(Equal(fmt.Sprintf(`  Generating SBOM for %s
      Completed in 0s

  Writing SBOM in the following format(s):
    application/vnd.cyclonedx+json
    application/spdx+json
    application/vnd.syft+json

`, path)))
		})
	})
}
