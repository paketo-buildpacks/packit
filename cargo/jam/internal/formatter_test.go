package internal_test

import (
	"bytes"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormatter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer    *bytes.Buffer
		formatter internal.Formatter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		formatter = internal.NewFormatter(buffer)
	})

	context("Markdown", func() {
		it("returns a list of dependencies", func() {
			formatter.Markdown([]cargo.ConfigMetadataDependency{
				{
					ID:      "some-dependency",
					Stacks:  []string{"some-stack"},
					Version: "1.2.3",
				},
				{
					ID:      "some-dependency",
					Stacks:  []string{"other-stack"},
					Version: "1.2.3",
				},
				{
					ID:      "other-dependency",
					Stacks:  []string{"some-stack", "other-stack"},
					Version: "2.3.4",
				},
				{
					ID:      "other-dependency",
					Stacks:  []string{"other-stack"},
					Version: "2.3.5",
				},
			}, map[string]string{
				"some-dependency":  "1.2.x",
				"other-dependency": "2.3.x",
			})
			Expect(buffer.String()).To(ContainSubstring(`Dependencies:
| name | version | stacks |
|-|-|-|
| other-dependency | 2.3.5 | other-stack |
| other-dependency | 2.3.4 | other-stack, some-stack |
| some-dependency | 1.2.3 | other-stack, some-stack |

Default dependencies:
| name | version |
|-|-|
| other-dependency | 2.3.x |
| some-dependency | 1.2.x |

Supported stacks:
| name |
|-|
| other-stack |
| some-stack |
`))
		})
	})
}
