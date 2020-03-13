// Package postal provides a service for resolving and installing dependencies
// for a buildpack.
//
// Below is an example that show the resolution and installation of a "node" dependency:
//
//   package main
//
//   import (
//   	"log"
//
//   	"github.com/cloudfoundry/packit/cargo"
//   	"github.com/cloudfoundry/packit/postal"
//   )
//
//   func main() {
//   	// Here we construct a transport and service so that we can download or fetch
//   	// dependencies from a cache and install them into a layer.
//   	transport := cargo.NewTransport()
//   	service := postal.NewService(transport)
//
//   	// The Resolve method can be used to pick a dependency that best matches a
//   	// set of criteria including id, version constraint, and stack.
//   	dependency, err := service.Resolve("/cnbs/com.example.nodejs-cnb/buildpack.toml", "node", "10.*", "com.example.stacks.bionic")
//   	if err != nil {
//   		log.Fatal(err)
//   	}
//
//   	// The Install method will download or fetch the given dependency and ensure
//   	// it is expanded into the given layer path as well as validated against its
//   	// SHA256 checksum.
//   	err = service.Install(dependency, "/cnbs/com.example.nodejs-cnb", "/layers/com.example.nodejs-cnb/node")
//   	if err != nil {
//   		log.Fatal(err)
//   	}
//   }
//
package postal
