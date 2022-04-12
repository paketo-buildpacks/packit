package planning_test

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/v2/planning"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testMetadata(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("Metadata", func() {
		it("will encode empty value to TOML as expected", func() {
			buffer := bytes.NewBuffer(nil)
			err := toml.NewEncoder(buffer).Encode(planning.Metadata{})
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(Equal(`version = ""
version-source = ""
build = false
launch = false
`))
		})

		it("will encode to TOML as expected", func() {
			buffer := bytes.NewBuffer(nil)
			err := toml.NewEncoder(buffer).Encode(planning.Metadata{
				Version:       "version-from-test",
				VersionSource: "version-source-from-test",
				Build:         true,
				Launch:        false,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(Equal(`version = "version-from-test"
version-source = "version-source-from-test"
build = true
launch = false
`))
		})

		it("will decode from nil as expected", func() {
			Expect(planning.NewMetadata(nil)).To(Equal(planning.Metadata{
				Version:       "",
				VersionSource: "",
				Build:         false,
				Launch:        false,
			}))
		})

		it("will decode from map as expected", func() {
			metadata := map[string]interface{}{
				"version":        "decoded-version",
				"version-source": "decoded-version-source",
				"build":          true,
				"launch":         true,
			}

			Expect(planning.NewMetadata(metadata)).To(Equal(planning.Metadata{
				Version:       "decoded-version",
				VersionSource: "decoded-version-source",
				Build:         true,
				Launch:        true,
			}))
		})

		it("does not decode invalid data", func() {
			metadata := map[string]interface{}{
				"version":        1,
				"version-source": 2,
				"build":          3,
				"launch":         4,
			}

			Expect(planning.NewMetadata(metadata)).To(Equal(planning.Metadata{}))
		})

		context("ToMap", func() {
			it("will translate as expected", func() {
				metadata := planning.Metadata{
					Version:       "version-from-test",
					VersionSource: "version-source-from-test",
					Build:         true,
					Launch:        false,
				}

				Expect(metadata.ToMap()).To(Equal(map[string]interface{}{
					"version":        "version-from-test",
					"version-source": "version-source-from-test",
					"build":          true,
					"launch":         false,
				}))
			})
		})
	})
}
