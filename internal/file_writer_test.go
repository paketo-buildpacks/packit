package internal_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/paketo-buildpacks/packit/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFileWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tmpDir     string
		path       string
		fileWriter internal.FileWriter
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "file-writer")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "file.ext")
	})

	it("writes the contents of a reader out to a file path", func() {
		err := fileWriter.Write(path, strings.NewReader("some-file-contents"))
		Expect(err).NotTo(HaveOccurred())

		contents, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("some-file-contents"))
	})

	context("failure cases", func() {
		context("when the file path cannot be created", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				err := fileWriter.Write(path, strings.NewReader("some-file-contents"))
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the reader throws an error", func() {
			it("returns an error", func() {
				err := fileWriter.Write(path, iotest.ErrReader(errors.New("failed to read")))
				Expect(err).To(MatchError("failed to read"))
			})
		})
	})
}
