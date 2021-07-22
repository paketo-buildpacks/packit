package extensions_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/extensions"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testServiceBindingsManager(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("ListServiceBindings", func() {
		var root string

		it.Before(func() {
			var err error
			root, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(root, "some-binding"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(root, "some-binding", "type"), []byte("some-type"), 0600)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(root, "some-binding", "provider"), []byte("some-provider"), 0600)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(root, "some-binding", "username"), []byte("some-username"), 0600)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(root, "some-binding", "password"), []byte("some-password"), 0600)
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(root, "other-binding"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(root, "other-binding", "type"), []byte("other-type"), 0600)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(root, "other-binding", "connection-count"), []byte("other-connection-count"), 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(root)).To(Succeed())
		})

		it("returns a list of all the platform bindings", func() {
			bindings, err := extensions.ListServiceBindings(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(bindings).To(HaveLen(2))
			Expect(bindings).To(ConsistOf([]extensions.ServiceBinding{
				{
					Name:     "some-binding",
					Type:     "some-type",
					Provider: "some-provider",
					Path:     filepath.Join(root, "some-binding"),
					Secrets: map[string][]byte{
						"username": []byte("some-username"),
						"password": []byte("some-password"),
					},
				},
				{
					Name: "other-binding",
					Type: "other-type",
					Path: filepath.Join(root, "other-binding"),
					Secrets: map[string][]byte{
						"connection-count": []byte("other-connection-count"),
					},
				},
			}))
		})

		context("failure cases", func() {
			context("when a secret cannot be read", func() {
				it.Before(func() {
					err := os.WriteFile(filepath.Join(root, "other-binding", "password"), []byte("other-password"), 0000)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := extensions.ListServiceBindings(root)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
