package vacation_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dsnetBzip2 "github.com/dsnet/compress/bzip2"
	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"
	"github.com/ulikunitz/xz"

	. "github.com/onsi/gomega"
)

func testArchive(t *testing.T, context spec.G, it spec.S) {
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

		context("when passed the reader of a bzip2 file", func() {
			var (
				archive vacation.Archive
				tempDir string
			)

			it.Before(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "vacation")
				Expect(err).NotTo(HaveOccurred())

				buffer := bytes.NewBuffer(nil)

				// Using the dsnet library because the Go compression library does not
				// have a writer. There is recent discussion on this issue
				// https://github.com/golang/go/issues/4828 to add an encoder. The
				// library should be removed once there is a native encoder
				bz, err := dsnetBzip2.NewWriter(buffer, nil)
				Expect(err).NotTo(HaveOccurred())

				tw := tar.NewWriter(bz)

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
				Expect(bz.Close()).To(Succeed())

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

				_, err = zw.Create("some-dir/")
				Expect(err).NotTo(HaveOccurred())

				header = &zip.FileHeader{Name: filepath.Join("some-dir", "some-nested-file")}
				header.SetMode(0644)

				f, err = zw.CreateHeader(header)
				Expect(err).NotTo(HaveOccurred())

				_, err = f.Write([]byte("nested file"))
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
					filepath.Join(tempDir, "some-dir"),
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

		context("when passed the reader of an executable file", func() {
			var (
				archive vacation.Archive
				tempDir string
				// Encoding of a very small elf executable from https://github.com/mathiasbynens/small
				encodedContents = []byte(`f0VMRgEBAQAAAAAAAAAAAAIAAwABAAAAGUDNgCwAAAAAAAAAAAAAADQAIAABAAAAAAAAAABAzYAAQM2ATAAAAEwAAAAFAAAAABAAAA==`)
				literalContents []byte
				fileName        = "exe"
			)

			it.Before(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "vacation")
				Expect(err).NotTo(HaveOccurred())

				literalContents, err = ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(encodedContents)))
				Expect(err).NotTo(HaveOccurred())

				archive = vacation.NewArchive(bytes.NewBuffer(literalContents)).WithName(fileName)
			})

			it.After(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			it("writes the executable in the bin dir", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(tempDir, "bin", fileName))
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(Equal(literalContents))
			})

			it("gives the executable execute permission", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				info, err := os.Stat(filepath.Join(tempDir, "bin", fileName))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(fs.FileMode(0755)))
			})
		})

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

			context("when given a name", func() {
				it.Before(func() {
					archive = archive.WithName("some-text-file")
				})

				it("writes a text file onto the path with that name", func() {
					err := archive.Decompress(tempDir)
					Expect(err).NotTo(HaveOccurred())

					content, err := os.ReadFile(filepath.Join(tempDir, "some-text-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(content).To(Equal([]byte(`some contents`)))
				})
			})
		})

		context("when passed the reader of a jar file", func() {
			var (
				archive vacation.Archive
				tempDir string
				header  []byte
			)

			it.Before(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "vacation")
				Expect(err).NotTo(HaveOccurred())

				// JAR header copied from https://github.com/gabriel-vasile/mimetype/blob/c4c6791c993e7f509de8ef38f149a59533e30bbc/testdata/jar.jar
				header = []byte("\x50\x4b\x03\x04\x14\x00\x08\x08\x08\x00\x59\x71\xbf\x4c\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x09\x00\x04\x00\x4d\x45\x54\x41\x2d\x49\x4e\x46\x2f\xfe\xca\x00\x00\x03\x00\x50\x4b\x07\x08\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x50\x4b\x03\x04\x14\x00\x08\x08\x08\x00\x59\x71\xbf\x4c\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x14\x00\x00\x00\x4d\x45\x54\x41\x2d\x49\x4e\x46\x2f\x4d\x41\x4e\x49\x46\x45\x53\x54\x2e\x4d\x46\xf3\x4d\xcc\xcb\x4c\x4b\x2d\x2e\xd1\x0d\x4b\x2d\x2a\xce\xcc\xcf\xb3\x52\x30\xd4\x33\xe0\xe5\x72\x2e\x4a\x4d\x2c\x49\x4d\xd1\x75\xaa\x04\x09\x58\xe8\x19\xc4\x1b\x9a\x1a\x2a\x68\xf8\x17\x25\x26\xe7\xa4\x2a\x38\xe7\x17\x15\xe4\x17\x25\x96\x00\xd5\x6b\xf2\x72\xf9\x26\x66\xe6\xe9\x3a\xe7\x24\x16\x17\x5b\x29\x78\xa4\xe6\xe4\xe4\x87\xe7\x17\xe5\xa4\xf0\x72\xf1\x72\x01\x00\x50\x4b\x07\x08\x86\x7d\x5d\xeb\x5c\x00\x00\x00\x5d\x00\x00\x00\x50\x4b\x03\x04\x14\x00\x08\x08\x08\x00\x12\x71\xbf\x4c\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x10\x00\x00\x00\x48\x65\x6c\x6c\x6f\x57\x6f\x72\x6c\x64\x2e\x63\x6c\x61\x73\x73\x6d\x50\x4d\x4b\xc3\x40\x10\x7d\xdb\xa6\x4d\x13\x53\x53\x5b\x53\x3f\x0b\xf6\x50\x88\x22\xe6\xe2\xad\xe2\x45\x10\x0f\x45\x85\x88\x1e\x3c\x6d\xda\xa5\x6c\xd9\x24\x12\x13\xc1\x9f\xa5\x07\x05\x0f\xfe\x00\x7f\x94\x38\x1b\x85\x20\x74\x0f\xb3\x3b\x6f\xde\x9b\x79\xb3\x5f\xdf\x1f\x9f\x00\x8e\x31\xb0\xd1\x84\x6b\xa1\x83\xb5\x16\xba\x36\x7a\x58\x37\xe1\x99\xe8\x33\x34\x4f\x64\x22\xf3\x53\x86\xba\xbf\x7f\xcb\x60\x9c\xa5\x33\xc1\xe0\x4e\x64\x22\x2e\x8b\x38\x12\xd9\x0d\x8f\x14\x21\x46\xcc\x65\xc2\xd0\xf7\xef\x27\x0b\xfe\xc4\x03\xc5\x93\x79\x10\xe6\x99\x4c\xe6\x63\x2d\xb4\xc3\xb4\xc8\xa6\xe2\x5c\x6a\xb2\x7b\x21\x94\x4a\xef\xd2\x4c\xcd\x8e\x34\xdb\x81\x89\x96\x89\x0d\x07\x9b\xd8\x62\x68\x97\xe5\xc3\xbd\x92\x30\x34\xb1\xed\x60\x07\xbb\xd4\xa3\x92\x31\x74\xaa\x31\x57\xd1\x42\x4c\xf3\x7f\x50\xf8\xfc\x98\x8b\x98\x5c\xa7\x05\x15\xbc\x5f\x4f\x32\x0d\xae\xc9\x50\x4e\xb6\x04\x8f\xc7\x0c\xbd\x25\x30\x83\xf9\xa0\x33\x45\xdb\x78\xfe\xb2\x65\x30\x44\x83\xfe\x4b\x9f\x1a\x98\xb6\x4e\xd1\xa2\x6c\x40\x37\xa3\xbb\x71\xf0\x0e\xf6\x42\x0f\xb2\x4c\xb1\x59\x82\x9a\xb2\x02\xe7\x8f\x3a\x2a\xa5\x80\xf5\x8a\x5a\xb7\xfe\x06\xa3\xa2\xdb\x54\xa2\x1e\xd4\x55\x0b\xdb\xe5\x94\xd5\x1f\x50\x4b\x07\x08\xe5\x38\x99\x3f\x21\x01\x00\x00\xab\x01\x00\x00\x50\x4b\x01\x02\x14\x00\x14\x00\x08\x08\x08\x00\x59\x71\xbf\x4c\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x09\x00\x04\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x4d\x45\x54\x41\x2d\x49\x4e\x46\x2f\xfe\xca\x00\x00\x50\x4b\x01\x02\x14\x00\x14\x00\x08\x08\x08\x00\x59\x71\xbf\x4c\x86\x7d\x5d\xeb\x5c\x00\x00\x00\x5d\x00\x00\x00\x14\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x3d\x00\x00\x00\x4d\x45\x54\x41\x2d\x49\x4e\x46\x2f\x4d\x41\x4e\x49\x46\x45\x53\x54\x2e\x4d\x46\x50\x4b\x01\x02\x14\x00\x14\x00\x08\x08\x08\x00\x12\x71\xbf\x4c\xe5\x38\x99\x3f\x21\x01\x00\x00\xab\x01\x00\x00\x10\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xdb\x00\x00\x00\x48\x65\x6c\x6c\x6f\x57\x6f\x72\x6c\x64\x2e\x63\x6c\x61\x73\x73\x50\x4b\x05\x06\x00\x00\x00\x00\x03\x00\x03\x00\xbb\x00\x00\x00\x3a\x02\x00\x00\x00\x00")
				buffer := bytes.NewBuffer(header)

				archive = vacation.NewArchive(buffer)
			})

			it.After(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			it("writes a jar file onto the path", func() {
				err := archive.Decompress(tempDir)
				Expect(err).NotTo(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(tempDir, "artifact"))
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(Equal(header))
			})

			context("when given a name", func() {
				it.Before(func() {
					archive = archive.WithName("some-jar-file")
				})

				it("writes a jar file onto the path with that name", func() {
					err := archive.Decompress(tempDir)
					Expect(err).NotTo(HaveOccurred())

					content, err := os.ReadFile(filepath.Join(tempDir, "some-jar-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(content).To(Equal(header))
				})
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
