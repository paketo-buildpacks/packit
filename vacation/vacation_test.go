package vacation_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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

	context("NewTarReader", func() {
		var (
			inputReader *bytes.Reader
		)

		it.Before(func() {
			inputReader = bytes.NewReader(nil)
		})

		context("when new tar reader is called", func() {
			it("returns a TarReadyReader object", func() {
				reader, err := vacation.NewTarReader(inputReader)
				Expect(err).ToNot(HaveOccurred())

				bufFinal := bytes.NewBuffer(nil)
				_, err = io.Copy(bufFinal, reader)
				Expect(err).ToNot(HaveOccurred())

				bufCompare := bytes.NewBuffer(nil)
				_, err = io.Copy(bufCompare, reader)
				Expect(err).ToNot(HaveOccurred())

				Expect(bufFinal).To(Equal(bufCompare))
			})
		})
	})

	context("NewGzipTarReader", func() {
		context("when the archive format is gzip tar", func() {
			var gzipReader *bytes.Reader

			it.Before(func() {
				var err error
				buffer := bytes.NewBuffer(nil)
				gzipWriter := gzip.NewWriter(buffer)
				tw := tar.NewWriter(gzipWriter)

				Expect(tw.WriteHeader(&tar.Header{Name: "some-file", Mode: 0755, Size: int64(len("some-file"))})).To(Succeed())
				_, err = tw.Write([]byte("some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(gzipWriter.Close()).To(Succeed())
				Expect(tw.Close()).To(Succeed())

				gzipReader = bytes.NewReader(buffer.Bytes())
			})

			it("returns a GZTarExtractor object", func() {
				reader, err := vacation.NewGzipTarReader(gzipReader)
				Expect(err).ToNot(HaveOccurred())

				bufFinal := bytes.NewBuffer(nil)
				_, err = io.Copy(bufFinal, reader)
				Expect(err).ToNot(HaveOccurred())

				_, err = gzipReader.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				gzipResults, err := gzip.NewReader(gzipReader)
				Expect(err).ToNot(HaveOccurred())

				bufCompare := bytes.NewBuffer(nil)
				_, err = io.Copy(bufCompare, gzipResults)
				Expect(err).ToNot(HaveOccurred())

				Expect(bufFinal.Bytes()).To(Equal(bufCompare.Bytes()))
			})

			context("failure case", func() {
				it("returns an error", func() {
					_, err := vacation.NewGzipTarReader(bytes.NewBuffer([]byte(`something`)))
					Expect(err).To(MatchError(ContainSubstring("failed to create gzip reader")))
				})
			})
		})

	})

	context("TarReadyReader.Decompress", func() {
		var (
			tempDir        string
			tarReadyReader vacation.TarReadyReader
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

			tarReadyReader, err = vacation.NewTarReader(bytes.NewReader(buffer.Bytes()))
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.Close()).To(Succeed())

		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			var err error
			err = tarReadyReader.Decompress(tempDir)
			Expect(err).ToNot(HaveOccurred())

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tempDir, "first"),
				filepath.Join(tempDir, "second"),
				filepath.Join(tempDir, "third"),
				filepath.Join(tempDir, "some-dir"),
			}))

			info, err := os.Stat(filepath.Join(tempDir, "first"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			Expect(filepath.Join(tempDir, "some-dir", "some-other-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "some-dir", "some-other-dir", "some-file")).To(BeARegularFile())
		})

		context("failure cases", func() {
			context("when it fails to read the tar response", func() {
				it("returns an error", func() {
					readyReader, err := vacation.NewTarReader(bytes.NewBuffer([]byte(`something`)))
					Expect(err).NotTo(HaveOccurred())

					err = readyReader.Decompress(tempDir)
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
					err := tarReadyReader.Decompress(tempDir)
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
					err := tarReadyReader.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to create archived file")))
				})
			})
		})
	})
}
