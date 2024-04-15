package servicebindings_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEntry(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect         = NewWithT(t).Expect
		entry          *servicebindings.Entry
		entryWithValue *servicebindings.Entry
		tmpDir         string
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "entry")
		Expect(err).NotTo(HaveOccurred())
		entryPath := filepath.Join(tmpDir, "entry")
		Expect(os.WriteFile(entryPath, []byte("some data"), os.ModePerm)).To(Succeed())
		entry = servicebindings.NewEntry(entryPath)
		entryWithValue = servicebindings.NewWithValue([]byte("value from env"))
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	context("ReadBytes", func() {
		it("returns the raw bytes of the entry", func() {
			Expect(entry.ReadBytes()).To(Equal([]byte("some data")))
			Expect(entryWithValue.ReadBytes()).To(Equal([]byte("value from env")))
		})
	})

	context("ReadString", func() {
		it("returns the string value of the entry", func() {
			Expect(entry.ReadString()).To(Equal("some data"))
			Expect(entryWithValue.ReadString()).To(Equal("value from env"))
		})
	})

	context("usage as an io.ReadCloser", func() {
		it("is assignable to io.ReadCloser", func() {
			var _ io.ReadCloser = entry
		})

		it("can be read again after closing", func() {
			data, err := io.ReadAll(entry)
			Expect(err).NotTo(HaveOccurred())
			Expect(entry.Close()).To(Succeed())
			Expect(data).To(Equal([]byte("some data")))

			data, err = io.ReadAll(entry)
			Expect(err).NotTo(HaveOccurred())
			Expect(entry.Close()).To(Succeed())
			Expect(data).To(Equal([]byte("some data")))

			data, err = io.ReadAll(entryWithValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(entryWithValue.Close()).To(Succeed())
			Expect(data).To(Equal([]byte("value from env")))

			data, err = io.ReadAll(entryWithValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(entryWithValue.Close()).To(Succeed())
			Expect(data).To(Equal([]byte("value from env")))
		})

		it("can be closed multiple times in a row", func() {
			_, err := io.ReadAll(entry)
			Expect(err).NotTo(HaveOccurred())
			Expect(entry.Close()).To(Succeed())
			Expect(entry.Close()).To(Succeed())

			_, err = io.ReadAll(entryWithValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(entryWithValue.Close()).To(Succeed())
			Expect(entryWithValue.Close()).To(Succeed())
		})

		it("can be closed if never read from", func() {
			Expect(entry.Close()).To(Succeed())
			Expect(entryWithValue.Close()).To(Succeed())
		})
	})
}
