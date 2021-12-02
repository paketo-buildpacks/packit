package vacation_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"
	"github.com/ulikunitz/xz"

	. "github.com/onsi/gomega"
)

func testXZArchive(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("Decompress", func() {
		var (
			tempDir   string
			xzArchive vacation.XZArchive
		)

		it.Before(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			xzw, err := xz.NewWriter(buffer)
			Expect(err).NotTo(HaveOccurred())

			tw := tar.NewWriter(xzw)

			Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
			Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
			_, err = tw.Write([]byte(nestedFile))
			Expect(err).NotTo(HaveOccurred())

			for _, file := range []string{"first", "second", "third"} {
				Expect(tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})).To(Succeed())
				_, err = tw.Write([]byte(file))
				Expect(err).NotTo(HaveOccurred())
			}

			Expect(tw.WriteHeader(&tar.Header{Name: "symlink", Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "first"})).To(Succeed())
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.Close()).To(Succeed())
			Expect(xzw.Close()).To(Succeed())

			xzArchive = vacation.NewXZArchive(bytes.NewReader(buffer.Bytes()))

		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("unpackages the archive into the path", func() {
			var err error
			err = xzArchive.Decompress(tempDir)
			Expect(err).ToNot(HaveOccurred())

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tempDir, "first"),
				filepath.Join(tempDir, "second"),
				filepath.Join(tempDir, "third"),
				filepath.Join(tempDir, "some-dir"),
				filepath.Join(tempDir, "symlink"),
			}))

			info, err := os.Stat(filepath.Join(tempDir, "first"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			Expect(filepath.Join(tempDir, "some-dir", "some-other-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "some-dir", "some-other-dir", "some-file")).To(BeARegularFile())

			data, err := os.ReadFile(filepath.Join(tempDir, "symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte(`first`)))
		})

		it("unpackages the archive into the path but also strips the first component", func() {
			var err error
			err = xzArchive.StripComponents(1).Decompress(tempDir)
			Expect(err).ToNot(HaveOccurred())

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tempDir, "some-other-dir"),
			}))

			Expect(filepath.Join(tempDir, "some-other-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "some-other-dir", "some-file")).To(BeARegularFile())

		})

		context("failure cases", func() {
			context("when it fails to create a xz reader", func() {
				it("returns an error", func() {
					readyArchive := vacation.NewXZArchive(bytes.NewBuffer([]byte(`something`)))

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create xz reader")))
				})
			})
		})
	})
}
