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

	. "github.com/onsi/gomega"
)

func testVacationTar(t *testing.T, context spec.G, it spec.S) {
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
			tempDir, err = os.MkdirTemp("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			tw := tar.NewWriter(buffer)

			// Some archive files will make a relative top level path directory these
			// should still successfully decompress.
			Expect(tw.WriteHeader(&tar.Header{Name: "./", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: "symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "first"})).To(Succeed())
			_, err = tw.Write([]byte{})
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

			Expect(tw.Close()).To(Succeed())

			tarArchive = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("unpackages the archive into the path", func() {
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

			data, err := os.ReadFile(filepath.Join(tempDir, "symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte(`first`)))
		})

		it("unpackages the archive into the path but also strips the first component", func() {
			var err error
			err = tarArchive.StripComponents(1).Decompress(tempDir)
			Expect(err).ToNot(HaveOccurred())

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tempDir, "some-other-dir"),
			}))

			Expect(filepath.Join(tempDir, "some-other-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "some-other-dir", "some-file")).To(BeARegularFile())

		})

		context("there is no directory metadata", func() {
			it.Before(func() {
				var err error

				buffer := bytes.NewBuffer(nil)
				tw := tar.NewWriter(buffer)

				nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
				Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
				_, err = tw.Write([]byte(nestedFile))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.WriteHeader(&tar.Header{Name: filepath.Join("sym-dir", "symlink"), Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: filepath.Join("..", nestedFile)})).To(Succeed())
				_, err = tw.Write([]byte{})
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.Close()).To(Succeed())

				tarArchive = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
			})

			it("unpackages the archive into the path", func() {
				err := tarArchive.Decompress(tempDir)
				Expect(err).ToNot(HaveOccurred())

				Expect(filepath.Join(tempDir, "some-dir", "some-other-dir")).To(BeADirectory())
				Expect(filepath.Join(tempDir, "some-dir", "some-other-dir", "some-file")).To(BeARegularFile())

				data, err := os.ReadFile(filepath.Join(tempDir, "sym-dir", "symlink"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal([]byte(filepath.Join("some-dir", "some-other-dir", "some-file"))))
			})
		})

		context("failure cases", func() {
			context("when a file is not inside of the destination director (Zip Slip)", func() {
				it.Before(func() {
					var err error

					buffer := bytes.NewBuffer(nil)
					tw := tar.NewWriter(buffer)

					nestedFile := filepath.Join("..", "some-dir", "some-other-dir", "some-file")
					Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
					_, err = tw.Write([]byte(nestedFile))
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())

					tarArchive = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
				})

				it("returns an error", func() {
					err := tarArchive.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("illegal file path %q: the file path does not occur within the destination directory", filepath.Join("..", "some-dir", "some-other-dir", "some-file")))))
				})
			})

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

				context("when there are no directory headers", func() {
					it.Before(func() {
						var err error

						buffer := bytes.NewBuffer(nil)
						tw := tar.NewWriter(buffer)

						nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
						Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
						_, err = tw.Write([]byte(nestedFile))
						Expect(err).NotTo(HaveOccurred())

						Expect(tw.Close()).To(Succeed())

						tarArchive = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
					})

					it("returns an error", func() {
						err := tarArchive.Decompress(tempDir)
						Expect(err).To(MatchError(ContainSubstring("failed to create archived directory from file path")))
					})
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

			context("when it tries to symlink to a file that does not exist", func() {
				var zipSlipSymlinkTar vacation.TarArchive

				it.Before(func() {
					var err error

					buffer := bytes.NewBuffer(nil)
					tw := tar.NewWriter(buffer)

					Expect(tw.WriteHeader(&tar.Header{Name: "symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: filepath.Join("..", "some-file")})).To(Succeed())
					_, err = tw.Write([]byte{})
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())

					zipSlipSymlinkTar = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
				})

				it("returns an error", func() {
					err := zipSlipSymlinkTar.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to evaluate symlink")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("when the symlink creation fails", func() {
				var brokenSymlinkTar vacation.TarArchive

				it.Before(func() {
					var err error

					buffer := bytes.NewBuffer(nil)
					tw := tar.NewWriter(buffer)

					Expect(tw.WriteHeader(&tar.Header{Name: "symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "some-file"})).To(Succeed())
					_, err = tw.Write([]byte{})
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())

					// Create a symlink in the target to force the new symlink create to
					// fail
					Expect(os.WriteFile(filepath.Join(tempDir, "some-file"), nil, 0644)).To(Succeed())
					Expect(os.Symlink("some-file", filepath.Join(tempDir, "symlink"))).To(Succeed())

					brokenSymlinkTar = vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
				})

				it("returns an error", func() {
					err := brokenSymlinkTar.Decompress(tempDir)
					Expect(err).To(MatchError(ContainSubstring("failed to extract symlink")))
				})
			})
		})
	})
}
