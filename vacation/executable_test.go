package vacation_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testExecutable(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Decompress", func() {
		var (
			archive vacation.Executable
			tempDir string
			// Encoding of a very small elf executable from https://github.com/mathiasbynens/small
			encodedContents = []byte(`f0VMRgEBAQAAAAAAAAAAAAIAAwABAAAAGUDNgCwAAAAAAAAAAAAAADQAIAABAAAAAAAAAABAzYAAQM2ATAAAAEwAAAAFAAAAABAAAA==`)
			literalContents []byte
		)

		it.Before(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			literalContents, err = io.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(encodedContents)))
			Expect(err).NotTo(HaveOccurred())

			archive = vacation.NewExecutable(bytes.NewBuffer(literalContents))
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		context("when passed the reader of an executable file", func() {
			it("writes the executable in the destination directory and sets the permissions using a default name", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(tempDir, "artifact"))
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(Equal(literalContents))

				info, err := os.Stat(filepath.Join(tempDir, "artifact"))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(fs.FileMode(0755)))
			})

			it("writes the executable in the destination directory and sets the permissions using a given name", func() {
				err := archive.WithName("executable").Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(tempDir, "executable"))
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(Equal(literalContents))

				info, err := os.Stat(filepath.Join(tempDir, "executable"))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(fs.FileMode(0755)))
			})
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
