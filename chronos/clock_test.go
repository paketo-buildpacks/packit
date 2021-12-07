package chronos_test

import (
	"errors"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testClock(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Now", func() {
		it("returns the value from the given Now function", func() {
			now := time.Now()

			clock := chronos.NewClock(func() time.Time {
				return now
			})

			Expect(clock.Now()).To(Equal(now))
		})
	})

	context("Measure", func() {
		var clock chronos.Clock

		it.Before(func() {
			now := time.Now()
			times := []time.Time{now, now.Add(20 * time.Second)}

			clock = chronos.NewClock(func() time.Time {
				t := time.Now()

				if len(times) > 0 {
					t = times[0]
					times = times[1:]
				}

				return t
			})
		})

		it("returns the duration taken to perform the operation", func() {
			duration, err := clock.Measure(func() error {
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(duration).To(Equal(20 * time.Second))
		})

		context("when the operation errors", func() {
			it("returns that error", func() {
				_, err := clock.Measure(func() error {
					return errors.New("operation failed")
				})
				Expect(err).To(MatchError("operation failed"))
			})
		})
	})
}
