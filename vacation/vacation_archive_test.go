package vacation_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"
	"github.com/ulikunitz/xz"

	. "github.com/onsi/gomega"
)

func testVacationArchive(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("NewArchive", func() {
		context("when passed the reader of a tar file", func() {
			var (
				archive vacation.Archive
				tempDir string
			)

			it.Before(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "vacation")
				Expect(err).NotTo(HaveOccurred())

				buffer := bytes.NewBuffer(nil)
				tw := tar.NewWriter(buffer)

				Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
				_, err = tw.Write(nil)
				Expect(err).NotTo(HaveOccurred())

				nestedFile := filepath.Join("some-dir", "some-nested-file")
				Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
				_, err = tw.Write([]byte(nestedFile))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.WriteHeader(&tar.Header{Name: "some-file", Mode: 0755, Size: int64(len("some-file"))})).To(Succeed())
				_, err = tw.Write([]byte("some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.Close()).To(Succeed())

				archive = vacation.NewArchive(buffer)
			})

			it.After(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			it("unpackages the archive into the path", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-dir"),
					filepath.Join(tempDir, "some-file"),
				}))
			})

			it("unpackages the archive into the path but also strips the first component", func() {
				err := archive.StripComponents(1).Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-nested-file"),
				}))
			})
		})

		context("when passed the reader of a tar gzip file", func() {
			var (
				archive vacation.Archive
				tempDir string
			)

			it.Before(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "vacation")
				Expect(err).NotTo(HaveOccurred())

				buffer := bytes.NewBuffer(nil)
				gw := gzip.NewWriter(buffer)
				tw := tar.NewWriter(gw)

				Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
				_, err = tw.Write(nil)
				Expect(err).NotTo(HaveOccurred())

				nestedFile := filepath.Join("some-dir", "some-nested-file")
				Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
				_, err = tw.Write([]byte(nestedFile))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.WriteHeader(&tar.Header{Name: "some-file", Mode: 0755, Size: int64(len("some-file"))})).To(Succeed())
				_, err = tw.Write([]byte("some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.Close()).To(Succeed())
				Expect(gw.Close()).To(Succeed())

				archive = vacation.NewArchive(buffer)
			})

			it.After(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			it("unpackages the archive into the path", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-dir"),
					filepath.Join(tempDir, "some-file"),
				}))
			})

			it("unpackages the archive into the path but also strips the first component", func() {
				err := archive.StripComponents(1).Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-nested-file"),
				}))
			})
		})

		context("when passed the reader of a tar xz file", func() {
			var (
				archive vacation.Archive
				tempDir string
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

				nestedFile := filepath.Join("some-dir", "some-nested-file")
				Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
				_, err = tw.Write([]byte(nestedFile))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.WriteHeader(&tar.Header{Name: "some-file", Mode: 0755, Size: int64(len("some-file"))})).To(Succeed())
				_, err = tw.Write([]byte("some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.Close()).To(Succeed())
				Expect(xzw.Close()).To(Succeed())

				archive = vacation.NewArchive(buffer)
			})

			it.After(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			it("unpackages the archive into the path", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-dir"),
					filepath.Join(tempDir, "some-file"),
				}))
			})

			it("unpackages the archive into the path but also strips the first component", func() {
				err := archive.StripComponents(1).Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-nested-file"),
				}))
			})
		})

		context("when passed the reader of a zip file", func() {
			var (
				archive vacation.Archive
				tempDir string
			)

			it.Before(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "vacation")
				Expect(err).NotTo(HaveOccurred())

				buffer := bytes.NewBuffer(nil)
				zw := zip.NewWriter(buffer)

				header := &zip.FileHeader{Name: "some-file"}
				header.SetMode(0755)

				f, err := zw.CreateHeader(header)
				Expect(err).NotTo(HaveOccurred())

				_, err = f.Write([]byte("some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(zw.Close()).To(Succeed())

				archive = vacation.NewArchive(buffer)
			})

			it.After(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			it("unpackages the archive into the path", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(tempDir, "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(tempDir, "some-file"),
				}))
			})
		})

		context("failure cases", func() {
			context("the buffer passed is of are unknown type", func() {
				var (
					archive vacation.Archive
					tempDir string
				)

				it.Before(func() {
					var err error
					tempDir, err = os.MkdirTemp("", "vacation")
					Expect(err).NotTo(HaveOccurred())

					// This is a FLAC header
					buffer := bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))

					archive = vacation.NewArchive(buffer)
				})

				it("returns an error", func() {
					err := archive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("unsupported archive type:")))
				})
			})
		})
	})
}
