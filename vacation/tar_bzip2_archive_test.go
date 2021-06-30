package vacation_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	dsnetBzip2 "github.com/dsnet/compress/bzip2"
	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testTarBzip2Archive(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("Decompress", func() {
		var (
			tempDir         string
			tarBzip2Archive vacation.TarBzip2Archive
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
			Expect(bz.Close()).To(Succeed())

			tarBzip2Archive = vacation.NewTarBzip2Archive(bytes.NewReader(buffer.Bytes()))
		})

		it.After(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		it("unpackages the archive into the path", func() {
			var err error
			err = tarBzip2Archive.Decompress(tempDir)
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
			err = tarBzip2Archive.StripComponents(1).Decompress(tempDir)
			Expect(err).ToNot(HaveOccurred())

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tempDir, "some-other-dir"),
			}))

			Expect(filepath.Join(tempDir, "some-other-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "some-other-dir", "some-file")).To(BeARegularFile())
		})
	})
}
