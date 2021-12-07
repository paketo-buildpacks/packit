package vacation_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testNopArchive(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Decompress", func() {
		var (
			archive vacation.NopArchive
			tempDir string
		)

		it.Before(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer([]byte(`some contents`))

			archive = vacation.NewNopArchive(buffer)
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("copies the contents of the reader to the destination with a default name", func() {
			err := archive.Decompress(filepath.Join(tempDir))
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(filepath.Join(tempDir, "artifact"))
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte(`some contents`)))
		})

		it("copies the contents of the reader to the destination with a given name", func() {
			err := archive.WithName("some-file").Decompress(filepath.Join(tempDir))
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(filepath.Join(tempDir, "some-file"))
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte(`some contents`)))
		})

		context("failure cases", func() {
			context("when the destination file cannot be created", func() {
				it("returns an error", func() {
					err := archive.Decompress("/no/such/path")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
}
