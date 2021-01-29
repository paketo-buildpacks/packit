package postal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDependencyMappings(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect      = NewWithT(t).Expect
		path        string
		dependency  postal.Dependency
		bindingPath string
		err         error
	)

	it.Before(func() {
		dependency = postal.Dependency{
			ID:      "some-entry",
			Stacks:  []string{"some-stack"},
			URI:     "some-uri",
			SHA256:  "some-sha",
			Version: "1.2.3",
		}

		bindingPath, err = ioutil.TempDir("", "bindings")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("FindDependencyMapping", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(bindingPath, "some-binding"), 0700)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(bindingPath, "some-binding", "type"), []byte("dependency-mapping"), 0600)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(bindingPath, "some-binding", "some-sha"), []byte("dependency-mapping-entry.tgz"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(bindingPath, "other-binding"), 0700)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(bindingPath, "other-binding", "type"), []byte("dependency-mapping"), 0600)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(bindingPath, "other-binding", "other-sha"), []byte("dependency-mapping-entry.tgz"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(bindingPath, "another-binding"), 0700)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(bindingPath, "another-binding", "type"), []byte("another type"), 0600)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(bindingPath, "another-binding", "some-sha"), []byte("entry.tgz"), 0600)).To(Succeed())
		})

		context("given a set of bindings and a dependency", func() {
			it("finds a matching dependency mappings in the platform bindings if there is one", func() {
				boundDependency, err := dependency.FindDependencyMapping(bindingPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(boundDependency).To(Equal(postal.DependencyMapping{
					SHA256: "some-sha",
					URI:    "dependency-mapping-entry.tgz",
				}))
			})
		})

		context("given a set of bindings and a dependency", func() {
			it.Before(func() {
				dependency.SHA256 = "unmatched-sha"
			})
			it("returns an empty DependencyMapping if there is no match", func() {
				boundDependency, err := dependency.FindDependencyMapping(bindingPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(boundDependency).To(Equal(postal.DependencyMapping{
					SHA256: "",
					URI:    "",
				}))
			})
		})
	})

	context("failure cases", func() {
		context("when the binding path is a bad pattern", func() {
			it("errors", func() {
				_, err := dependency.FindDependencyMapping("///")
				Expect(err).To(HaveOccurred())
			})
		})

		context("when type file cannot be opened", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(bindingPath, "some-binding"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(bindingPath, "some-binding", "type"), []byte("dependency-mapping"), 0000)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(bindingPath, "some-binding", "some-sha"), []byte("dependency-mapping-entry.tgz"), 0600)).To(Succeed())
			})
			it("errors", func() {
				_, err := dependency.FindDependencyMapping(bindingPath)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("couldn't read binding type")))
			})
		})

		context("when SHA256 file cannot be stat", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(bindingPath, "new-binding"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(bindingPath, "new-binding", "type"), []byte("dependency-mapping"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(bindingPath, "new-binding", "some-sha"), []byte("dependency-mapping-entry.tgz"), 0644)).To(Succeed())
				Expect(os.Chmod(filepath.Join(bindingPath, "new-binding", "some-sha"), 0000)).To(Succeed())
			})
			it("errors", func() {
				_, err := dependency.FindDependencyMapping(bindingPath)
				Expect(err).To(HaveOccurred())
			})
		})

		context("when SHA256 contents cannot be opened", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(bindingPath, "some-binding"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(bindingPath, "some-binding", "type"), []byte("dependency-mapping"), 0600)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(bindingPath, "some-binding", "some-sha"), []byte("dependency-mapping-entry.tgz"), 0000)).To(Succeed())
			})
			it("errors", func() {
				_, err := dependency.FindDependencyMapping(bindingPath)
				Expect(err).To(HaveOccurred())
			})
		})
	})
}
