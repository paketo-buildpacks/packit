package vacation_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSymlinkSorting(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("TarArchive test that symlinks are sorted so that symlink to other symlinks are created after the initial symlink", func() {
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

			Expect(tw.WriteHeader(&tar.Header{Name: "b-symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: filepath.Join("a-symlink", "x")})).To(Succeed())
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: "c-symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "d-symlink"})).To(Succeed())
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: "d-symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "z"})).To(Succeed())
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: "a-symlink", Mode: 0755, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "z"})).To(Succeed())
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.WriteHeader(&tar.Header{Name: "z", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			xFile := filepath.Join("z", "x")
			Expect(tw.WriteHeader(&tar.Header{Name: xFile, Mode: 0755, Size: int64(len(xFile))})).To(Succeed())
			_, err = tw.Write([]byte(xFile))
			Expect(err).NotTo(HaveOccurred())

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
				filepath.Join(tempDir, "a-symlink"),
				filepath.Join(tempDir, "b-symlink"),
				filepath.Join(tempDir, "z"),
				filepath.Join(tempDir, "d-symlink"),
				filepath.Join(tempDir, "c-symlink"),
			}))

			Expect(filepath.Join(tempDir, "z")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "z", "x")).To(BeARegularFile())

			link, err := os.Readlink(filepath.Join(tempDir, "a-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("z"))

			link, err = os.Readlink(filepath.Join(tempDir, "c-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("d-symlink"))

			link, err = os.Readlink(filepath.Join(tempDir, "d-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("z"))

			data, err := os.ReadFile(filepath.Join(tempDir, "b-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte(filepath.Join("z", "x"))))
		})
	})

	context("ZipArchive test that symlinks are sorted so that symlink to other symlinks are created after the initial symlink", func() {
		var (
			tempDir    string
			zipArchive vacation.ZipArchive
		)

		it.Before(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "vacation")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			zw := zip.NewWriter(buffer)

			fileHeader := &zip.FileHeader{Name: "b-symlink"}
			fileHeader.SetMode(0755 | os.ModeSymlink)

			bSymlink, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = bSymlink.Write([]byte(filepath.Join("a-symlink", "x")))
			Expect(err).NotTo(HaveOccurred())

			fileHeader = &zip.FileHeader{Name: "c-symlink"}
			fileHeader.SetMode(0755 | os.ModeSymlink)

			cSymlink, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = cSymlink.Write([]byte(`d-symlink`))
			Expect(err).NotTo(HaveOccurred())

			fileHeader = &zip.FileHeader{Name: "d-symlink"}
			fileHeader.SetMode(0755 | os.ModeSymlink)

			dSymlink, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = dSymlink.Write([]byte(`z`))
			Expect(err).NotTo(HaveOccurred())

			fileHeader = &zip.FileHeader{Name: "a-symlink"}
			fileHeader.SetMode(0755 | os.ModeSymlink)

			aSymlink, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = aSymlink.Write([]byte(`z`))
			Expect(err).NotTo(HaveOccurred())

			_, err = zw.Create("z" + string(filepath.Separator))
			Expect(err).NotTo(HaveOccurred())

			fileHeader = &zip.FileHeader{Name: filepath.Join("z", "x")}
			fileHeader.SetMode(0644)

			xFile, err := zw.CreateHeader(fileHeader)
			Expect(err).NotTo(HaveOccurred())

			_, err = xFile.Write([]byte("x file"))
			Expect(err).NotTo(HaveOccurred())

			Expect(zw.Close()).To(Succeed())

			zipArchive = vacation.NewZipArchive(bytes.NewReader(buffer.Bytes()))
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("unpackages the archive into the path", func() {
			var err error
			err = zipArchive.Decompress(tempDir)
			Expect(err).ToNot(HaveOccurred())

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tempDir, "a-symlink"),
				filepath.Join(tempDir, "b-symlink"),
				filepath.Join(tempDir, "z"),
				filepath.Join(tempDir, "d-symlink"),
				filepath.Join(tempDir, "c-symlink"),
			}))

			Expect(filepath.Join(tempDir, "z")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "z", "x")).To(BeARegularFile())

			link, err := os.Readlink(filepath.Join(tempDir, "a-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("z"))

			link, err = os.Readlink(filepath.Join(tempDir, "c-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("d-symlink"))

			link, err = os.Readlink(filepath.Join(tempDir, "d-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("z"))

			data, err := os.ReadFile(filepath.Join(tempDir, "b-symlink"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte(`x file`)))
		})
	})
}
