package cargo_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDependencyCacher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tmpDir     string
		downloader *fakes.Downloader
		cacher     cargo.DependencyCacher
	)

	it.Before(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "cacher-test")
		Expect(err).NotTo(HaveOccurred())

		downloader = &fakes.Downloader{}
		downloader.DropCall.Stub = func(root, uri string) (io.ReadCloser, error) {
			switch uri {
			case "http://dep1-uri":
				return ioutil.NopCloser(strings.NewReader("dep1-contents")), nil

			case "http://dep2-uri":
				return ioutil.NopCloser(strings.NewReader("dep2-contents")), nil

			case "http://error-dep":
				return ioutil.NopCloser(errorReader{}), nil

			default:
				return nil, fmt.Errorf("no such dependency: %s", uri)
			}
		}

		cacher = cargo.NewDependencyCacher(downloader)
	})

	context("Cache", func() {
		it("caches dependencies and returns updated dependencies list", func() {
			deps, err := cacher.Cache(tmpDir, []cargo.ConfigMetadataDependency{
				{
					URI:    "http://dep1-uri",
					SHA256: "3c9de6683673f3e8039599d5200d533807c6c35fd9e35d6b6d77009122868f0f",
				},
				{
					URI:    "http://dep2-uri",
					SHA256: "bfc72d62682f4a2edc3218d70b1f7052e4f336c179a8f19ef12ee721d4ea29b7",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(deps).To(Equal([]cargo.ConfigMetadataDependency{
				{
					URI:    "file:///dependencies/3c9de6683673f3e8039599d5200d533807c6c35fd9e35d6b6d77009122868f0f",
					SHA256: "3c9de6683673f3e8039599d5200d533807c6c35fd9e35d6b6d77009122868f0f",
				},
				{
					URI:    "file:///dependencies/bfc72d62682f4a2edc3218d70b1f7052e4f336c179a8f19ef12ee721d4ea29b7",
					SHA256: "bfc72d62682f4a2edc3218d70b1f7052e4f336c179a8f19ef12ee721d4ea29b7",
				},
			}))

			Expect(downloader.DropCall.Receives.Root).To(Equal(""))

			contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "dependencies", "3c9de6683673f3e8039599d5200d533807c6c35fd9e35d6b6d77009122868f0f"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("dep1-contents"))

			contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "dependencies", "bfc72d62682f4a2edc3218d70b1f7052e4f336c179a8f19ef12ee721d4ea29b7"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("dep2-contents"))
		})

		context("failure cases", func() {
			context("when the dependencies directory cannot be created", func() {
				it.Before(func() {
					Expect(os.Chmod(tmpDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tmpDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := cacher.Cache(tmpDir, nil)
					Expect(err).To(MatchError(ContainSubstring("failed to create dependencies directory:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when a dependency cannot be downloaded", func() {
				it("returns an error", func() {
					_, err := cacher.Cache(tmpDir, []cargo.ConfigMetadataDependency{
						{
							URI: "http://unknown-dep",
						},
					})
					Expect(err).To(MatchError("failed to download dependency: no such dependency: http://unknown-dep"))
				})
			})

			context("when the destination file cannot be created", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(tmpDir, "dependencies"), 0000)).To(Succeed())
				})

				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(tmpDir, "dependencies"), os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := cacher.Cache(tmpDir, []cargo.ConfigMetadataDependency{
						{
							URI:    "http://dep1-uri",
							SHA256: "3c9de6683673f3e8039599d5200d533807c6c35fd9e35d6b6d77009122868f0f",
						},
					})
					Expect(err).To(MatchError(ContainSubstring("failed to create destination file:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when we fail to read the downloaded file", func() {
				it("returns an error", func() {
					_, err := cacher.Cache(tmpDir, []cargo.ConfigMetadataDependency{
						{
							URI:    "http://error-dep",
							SHA256: "some-sha",
						},
					})
					Expect(err).To(MatchError("failed to copy dependency: failed to read"))
				})
			})

			context("when the checksum does not match", func() {
				it("returns an error", func() {
					_, err := cacher.Cache(tmpDir, []cargo.ConfigMetadataDependency{
						{
							URI:    "http://dep1-uri",
							SHA256: "invalid-sha",
						},
					})
					Expect(err).To(MatchError("failed to copy dependency: validation error: checksum does not match"))
				})
			})
		})
	})
}
