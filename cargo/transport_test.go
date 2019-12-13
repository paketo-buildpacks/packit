package cargo_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testTransport(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Drop", func() {
		var transport cargo.Transport

		it.Before(func() {
			transport = cargo.NewTransport()
		})

		context("when the given uri is online", func() {
			var server *httptest.Server

			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.URL.Path {
					case "/some-bundle":
						w.Write([]byte("some-bundle-contents"))
					default:
						http.NotFound(w, req)
					}
				}))
			})

			it.After(func() {
				server.Close()
			})

			it("downloads the file from a URI", func() {
				bundle, err := transport.Drop("", fmt.Sprintf("%s/some-bundle", server.URL))
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadAll(bundle)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-bundle-contents"))

				Expect(bundle.Close()).To(Succeed())
			})

			context("failure cases", func() {
				context("when the uri is malformed", func() {
					it("returns an error", func() {
						_, err := transport.Drop("", "%%%%")
						Expect(err).To(MatchError(ContainSubstring("failed to parse request uri")))
						Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
					})
				})

				context("when the request fails", func() {
					it.Before(func() {
						server.Close()
					})

					it("returns an error", func() {
						_, err := transport.Drop("", fmt.Sprintf("%s/some-bundle", server.URL))
						Expect(err).To(MatchError(ContainSubstring("failed to make request")))
						Expect(err).To(MatchError(ContainSubstring("connection refused")))
					})
				})
			})
		})

		context("when the uri is for a file", func() {
			var (
				path string
				dir  string
			)

			it.Before(func() {
				var err error
				dir, err = ioutil.TempDir("", "bundle")
				Expect(err).NotTo(HaveOccurred())

				path = "some-file"

				err = ioutil.WriteFile(filepath.Join(dir, path), []byte("some-bundle-contents"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				Expect(os.RemoveAll(dir)).To(Succeed())
			})

			it("returns the file descriptor", func() {
				bundle, err := transport.Drop(dir, fmt.Sprintf("file://%s", path))
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadAll(bundle)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-bundle-contents"))

				Expect(bundle.Close()).To(Succeed())
			})

			context("failure cases", func() {
				it.Before(func() {
					Expect(os.RemoveAll(dir)).To(Succeed())
				})

				context("when the file does not exist", func() {
					it("returns an error", func() {
						_, err := transport.Drop(dir, fmt.Sprintf("file://%s", path))
						Expect(err).To(MatchError(ContainSubstring("failed to open file")))
						Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
					})
				})
			})
		})
	})
}
