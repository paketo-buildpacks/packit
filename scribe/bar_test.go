package scribe_test

import (
	"bytes"
	"testing"

	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBar(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer *bytes.Buffer
		bar    *scribe.Bar
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)

		bar = scribe.NewBar(buffer)
	})

	it("renders a progress bar to the writer", func() {
		bar.Start()
		for i := 0; i < 40; i++ {
			bar.Increment()
		}
		bar.Finish()

		Expect(buffer.String()).To(Equal("[----------------------------------->                                                    ] 40.00% 0s"))
	})
}
