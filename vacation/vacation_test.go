package vacation_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/vacation"
	"github.com/sclevine/spec"
	"github.com/ulikunitz/xz"

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

	context("TarXZArchive.Decompress", func() {
		var (
			tempDir      string
			tarXZArchive vacation.TarXZArchive
		)

		it.Before(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "vacation")
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

			tarXZArchive = vacation.NewTarXZArchive(bytes.NewReader(buffer.Bytes()))

		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			var err error
			err = tarXZArchive.Decompress(tempDir)
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
			context("when it fails to create a xz reader", func() {
				it("returns an error", func() {
					readyArchive := vacation.NewTarXZArchive(bytes.NewBuffer([]byte(`something`)))

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create xz reader")))
				})
			})
		})
	})

	context("ZipArchive.Decompress", func() {
		var (
			tempDir    string
			zipArchive vacation.ZipArchive
		)

		it.Before(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			zw := zip.NewWriter(buffer)

			_, err = zw.Create("some-dir/")
			Expect(err).NotTo(HaveOccurred())

			_, err = zw.Create(fmt.Sprintf("%s/", filepath.Join("some-dir", "some-other-dir")))
			Expect(err).NotTo(HaveOccurred())

			fileHeader := &zip.FileHeader{Name: filepath.Join("some-dir", "some-other-dir", "some-file")}
			fileHeader.SetMode(0644)

			nestedFile, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = nestedFile.Write([]byte("nested file"))
			Expect(err).NotTo(HaveOccurred())

			for _, name := range []string{"first", "second", "third"} {
				fileHeader := &zip.FileHeader{Name: name}
				fileHeader.SetMode(0755)

				f, err := zw.CreateHeader(fileHeader)
				Expect(err).NotTo(HaveOccurred())

				_, err = f.Write([]byte(name))
				Expect(err).NotTo(HaveOccurred())
			}

			fileHeader = &zip.FileHeader{Name: "symlink"}
			fileHeader.SetMode(0755 | os.ModeSymlink)

			symlink, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = symlink.Write([]byte(filepath.Join("some-dir", "some-other-dir", "some-file")))
			Expect(err).NotTo(HaveOccurred())

			Expect(zw.Close()).To(Succeed())

			zipArchive = vacation.NewZipArchive(bytes.NewReader(buffer.Bytes()))
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			var err error
			err = zipArchive.Decompress(tempDir)
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

			link, err := os.Readlink(filepath.Join(tempDir, "symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("some-dir/some-other-dir/some-file"))

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte("nested file")))
		})

		context("failure cases", func() {
			context("when it fails to create a zip reader", func() {
				it("returns an error", func() {
					readyArchive := vacation.NewZipArchive(bytes.NewBuffer([]byte(`something`)))

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create zip reader")))
				})
			})

			context("when it fails to unzip a directory", func() {
				var buffer *bytes.Buffer
				it.Before(func() {
					var err error
					buffer = bytes.NewBuffer(nil)
					zw := zip.NewWriter(buffer)

					_, err = zw.Create("some-dir/")
					Expect(err).NotTo(HaveOccurred())

					Expect(zw.Close()).To(Succeed())

					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					readyArchive := vacation.NewZipArchive(buffer)

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to unzip directory")))
				})
			})

			context("when it fails to unzip a directory that is part of a file base", func() {
				var buffer *bytes.Buffer
				it.Before(func() {
					var err error
					buffer = bytes.NewBuffer(nil)
					zw := zip.NewWriter(buffer)

					_, err = zw.Create("some-dir/some-file")
					Expect(err).NotTo(HaveOccurred())

					Expect(zw.Close()).To(Succeed())

					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					readyArchive := vacation.NewZipArchive(buffer)

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to unzip directory that was part of file path")))
				})
			})

			context("when it fails to unzip a symlink", func() {
				var buffer *bytes.Buffer
				it.Before(func() {
					var err error
					buffer = bytes.NewBuffer(nil)
					zw := zip.NewWriter(buffer)

					header := &zip.FileHeader{Name: "symlink"}
					header.SetMode(0755 | os.ModeSymlink)

					symlink, err := zw.CreateHeader(header)
					Expect(err).NotTo(HaveOccurred())

					_, err = symlink.Write([]byte(filepath.Join("some", "path", "to", "a", "target")))
					Expect(err).NotTo(HaveOccurred())

					Expect(zw.Close()).To(Succeed())

					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					readyArchive := vacation.NewZipArchive(buffer)

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to unzip symlink")))
				})
			})

			context("when it fails to unzip a file", func() {
				var buffer *bytes.Buffer
				it.Before(func() {
					var err error
					buffer = bytes.NewBuffer(nil)
					zw := zip.NewWriter(buffer)

					_, err = zw.Create("some-file")
					Expect(err).NotTo(HaveOccurred())

					Expect(zw.Close()).To(Succeed())

					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					readyArchive := vacation.NewZipArchive(buffer)

					err := readyArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to unzip file")))
				})
			})
		})
	})
}
