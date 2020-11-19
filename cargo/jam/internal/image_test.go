package internal_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testImage(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		server *httptest.Server
	)

	context("FindLatestImage", func() {
		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				switch req.URL.Path {
				case "/v2/":
					w.WriteHeader(http.StatusOK)

				case "/v2/some-org/some-repo/tags/list":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{
						  "tags": [
								"0.0.10",
								"0.20.1",
								"0.20.12",
								"latest"
							]
					}`)

				case "/v2/some-org/error-repo/tags/list":
					w.WriteHeader(http.StatusTeapot)

				default:
					t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
				}
			}))
		})

		it("returns the latest semver tag for the given image uri", func() {
			image, err := internal.FindLatestImage(fmt.Sprintf("%s/some-org/some-repo:latest", strings.TrimPrefix(server.URL, "http://")))
			Expect(err).NotTo(HaveOccurred())
			Expect(image).To(Equal(internal.Image{
				Name:    fmt.Sprintf("%s/some-org/some-repo", strings.TrimPrefix(server.URL, "http://")),
				Path:    "some-org/some-repo",
				Version: "0.20.12",
			}))
		})

		context("failure cases", func() {
			context("when the uri cannot be parsed", func() {
				it("returns an error", func() {
					_, err := internal.FindLatestImage("not a valid uri")
					Expect(err).To(MatchError("failed to parse image reference \"not a valid uri\": invalid reference format"))
				})
			})

			context("when the repo name cannot be parsed", func() {
				it("returns an error", func() {
					_, err := internal.FindLatestImage(fmt.Sprintf("%s/a:latest", strings.TrimPrefix(server.URL, "http://")))
					Expect(err).To(MatchError("failed to parse image repository: repository must be between 2 and 255 runes in length: a"))
				})
			})

			context("when the tags cannot be listed", func() {
				it("returns an error", func() {
					_, err := internal.FindLatestImage(fmt.Sprintf("%s/some-org/error-repo:latest", strings.TrimPrefix(server.URL, "http://")))
					Expect(err).To(MatchError(ContainSubstring("failed to list tags:")))
					Expect(err).To(MatchError(ContainSubstring("unsupported status code 418")))
				})
			})
		})
	})
}
