// Package chronos provides clock functionality that can be useful when
// developing and testing Cloud Native Buildpacks.
//
// Below is an example showing how you might use a Clock to measure the
// duration of an operation:
//
//   package main
//
//   import (
//   	"os"
//
//   	"github.com/paketo-buildpacks/packit/chronos"
//   )
//
//   func main() {
//   	duration, err := chronos.DefaultClock.Measure(func() error {
//      // Perform some operation, like sleep for 10 seconds
//      time.Sleep(10 * time.Second)
//
//      return nil
//    })
//   	if err != nil {
//   		panic(err)
//   	}
//
//    fmt.Printf("duration: %s", duration)
//   	// Output: duration: 10s
//   }
//
package chronos
