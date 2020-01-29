package vacation_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVacation(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("TarArchive.Decompress", func() {
		var (
			tempDir    string
			tarArchive vacation.TarArchive
		)

		it.Before(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			tw := tar.NewWriter(buffer)

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

			Expect(tw.WriteHeader(&tar.Header{Name: "symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "first"})).To(Succeed())
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.Close()).To(Succeed())

			tarArchive = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))

		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			var err error
			err = tarArchive.Decompress(tempDir)
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

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte(`first`)))
		})

		context("failure cases", func() {
			context("when it fails to read the tar response", func() {
				it("returns an error", func() {
					readyArchive := vacation.NewTarArchive(bytes.NewBuffer([]byte(`something`)))

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to read tar response")))

				})
			})

			context("when it is unable to create an archived directory", func() {
				it.Before(func() {
					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := tarArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create archived directory")))
				})
			})

			context("when it is unable to create an archived file", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(tempDir, "some-dir", "some-other-dir"), os.ModePerm)).To(Succeed())
					Expect(os.Chmod(filepath.Join(tempDir, "some-dir", "some-other-dir"), 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(filepath.Join(tempDir, "some-dir", "some-other-dir"), os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := tarArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create archived file")))
				})
			})

			context("when it tries to decompress a broken symlink", func() {
				var brokenSymlinkTar vacation.TarArchive

				it.Before(func() {
					var err error

					buffer := bytes.NewBuffer(nil)
					tw := tar.NewWriter(buffer)

					Expect(tw.WriteHeader(&tar.Header{Name: "symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: ""})).To(Succeed())
					_, err = tw.Write([]byte{})
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())

					brokenSymlinkTar = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
				})

				it("returns an error", func() {
					err := brokenSymlinkTar.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to extract symlink")))
				})
			})
		})
	})

	context("TarGzipArchive.Decompress", func() {
		var (
			tempDir        string
			tarGzipArchive vacation.TarGzipArchive
		)

		it.Before(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			gw := gzip.NewWriter(buffer)
			tw := tar.NewWriter(gw)

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
			Expect(gw.Close()).To(Succeed())

			tarGzipArchive = vacation.NewTarGzipArchive(bytes.NewReader(buffer.Bytes()))

		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			var err error
			err = tarGzipArchive.Decompress(tempDir)
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

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte(`first`)))
		})

		context("failure cases", func() {
			context("when it fails to create a grip reader", func() {
				it("returns an error", func() {
					readyArchive := vacation.NewTarGzipArchive(bytes.NewBuffer([]byte(`something`)))

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create gzip reader")))
				})
			})
		})
	})
}
