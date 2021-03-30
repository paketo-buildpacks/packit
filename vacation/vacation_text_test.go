package vacation_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVacationText(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("when passed the reader of a text file", func() {
		var (
			archive vacation.Archive
			tempDir string
		)

		it.Before(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer([]byte(`some contents`))

			archive = vacation.NewArchive(buffer)
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("writes a text file onto the path", func() {
			err := archive.Decompress(tempDir)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(filepath.Join(tempDir, "artifact"))
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte(`some contents`)))
		})
	})
}
