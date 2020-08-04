// Package packit provides primitives for implementing a Cloud Native Buildpack
// according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md.
//
// Buildpack Interface
//
// According to the specification, the buildpack interface is composed of both
// a detect and build phase. Each of these phases has a corresponding set of
// packit primitives enable developers to easily implement a buildpack.
//
// Detect Phase
//
// The purpose of the detect phase is for buildpacks to declare dependencies
// that are provided or required for the buildpack to execute. Implementing the
// detect phase can be achieved by calling the Detect function and providing a
// DetectFunc callback to be invoked during that phase. Below is an example of
// a simple detect phase that provides the "yarn" dependency and requires the
// "node" dependency.
//
//   package main
//
//   import (
//   	"encoding/json"
//   	"os"
//   	"path/filepath"
//
//   	"github.com/paketo-buildpacks/packit"
//   )
//
//   func main() {
//   	// The detect phase provides yarn and requires node. When requiring node,
//   	// a version specified in the package.json file is included to indicate
//   	// what versions of node are acceptable to the buildpack.
//
//   	packit.Detect(func(context packit.DetectContext) (packit.DetectResult, error) {
//
//   		// The DetectContext includes a WorkingDir field that specifies the
//   		// location of the application source code. This field can be combined with
//   		// other paths to find and inspect files included in the application source
//   		// code that is provided to the buildpack.
//   		file, err := os.Open(filepath.Join(context.WorkingDir, "package.json"))
//   		if err != nil {
//   			return packit.DetectResult{}, err
//   		}
//
//   		// The package.json file includes a declaration of what versions of node
//   		// are acceptable. For example:
//   		//   {
//   		//     "engines": {
//   		//       "node": ">=0.10.3 <0.12"
//   		//     }
//   		//   }
//   		var config struct {
//   			Engines struct {
//   				Node string `json:"node"`
//   			} `json:"engines"`
//   		}
//
//   		err = json.NewDecoder(file).Decode(&config)
//   		if err != nil {
//   			return packit.DetectResult{}, err
//   		}
//
//   		// Once the package.json file has been parsed, the detect phase can return
//   		// a result that indicates the provision of yarn and the requirement of
//   		// node. As can be seen below, the BuildPlanRequirement may also include
//   		// optional metadata information to such as the source of the version
//   		// information for a given requirement.
//   		return packit.DetectResult{
//   			Plan: packit.BuildPlan{
//   				Provides: []packit.BuildPlanProvision{
//   					{Name: "yarn"},
//   				},
//   				Requires: []packit.BuildPlanRequirement{
//   					{
//   						Name:    "node",
//   						Version: config.Engines.Node,
//   						Metadata: map[string]string{
//   							"version-source": "package.json",
//   						},
//   					},
//   				},
//   			},
//   		}, nil
//   	})
//   }
//
// Build Phase
//
// The purpose of the build phase is to perform the operation of providing
// whatever dependencies were declared in the detect phase for the given
// application code. Implementing the build phase can be achieved by calling
// the Build function and providing a BuildFunc callback to be invoked during
// that phase. Below is an example that adds "yarn" as a dependency to the
// application source code.
//
//   package main
//
//   import "github.com/paketo-buildpacks/packit"
//
//   func main() {
//   	// The build phase includes the yarn cli in a new layer that is made
//   	// available for subsequent buildpacks during their build phase as well as to
//   	// the start command during launch.
//
//   	packit.Build(func(context packit.BuildContext) (packit.BuildResult, error) {
//
//   		// The BuildContext includes a BuildpackPlan with entries that specify a
//   		// requirement on a dependency provided by the buildpack. This example
//   		// simply chooses the first entry, but more intelligent resolution
//   		// processes can and likely shoud be used in real implementations.
//   		entry := context.Plan.Entries[0]
//
//   		// The BuildContext also provides a mechanism whereby a layer can be
//   		// created to store the results of a given portion of the build process.
//   		// This example creates a layer called "yarn" that will hold the yarn cli.
//   		layer, err := context.Layers.Get("yarn", packit.BuildLayer, packit.LaunchLayer)
//   		if err != nil {
//   			return packit.BuildResult{}, err
//   		}
//
//   		// At this point we are performing the process of installing the yarn cli.
//   		// As those details are not important to the explanation of the packit API,
//   		// they are omitted here.
//   		err = InstallYarn(entry.Version, layer.Path)
//   		if err != nil {
//   			return packit.BuildResult{}, err
//   		}
//
//   		// After the installation of the yarn cli, a BuildResult can be returned
//   		// that included details of the executed BuildpackPlan, the Layers to
//   		// provide back to the lifecycle, and the Processes to execute at launch.
//   		return packit.BuildResult{
//   			Layers: []packit.Layer{
//   				layer,
//   			},
//   			Processes: []packit.Process{
//   				{
//   					Type:    "web",
//   					Command: "yarn start",
//   				},
//   			},
//   		}, nil
//   	})
//   }
//
//   // InstallYarn executes the process of installing the yarn cli.
//   func InstallYarn(version, path string) error {
//   	// Implemention omitted.
//   	return nil
//   }
//
// Run
//
// Buildpacks can be created with a single entrypoint executable using the
// packit.Run function. Here, you can combine both the Detect and Build phases
// and run will ensure that the correct phase is called when the matching
// executable is called by the Cloud Native Buildpack Lifecycle. Below is an
// example that combines a simple detect and build into a single main program.
//
//   package main
//
//   import "github.com/paketo-buildpacks/packit"
//
//   func main() {
//   	detect := func(context packit.DetectContext) (packit.DetectResult, error) {
//   		return packit.DetectResult{}, nil
//   	}
//
//   	build := func(context packit.BuildContext) (packit.BuildResult, error) {
//   		return packit.BuildResult{
//   			Processes: []packit.Process{
//   				{
//   					Type:    "web",
//   					Command: `while true; do nc -l -p $PORT -c 'echo -e "HTTP/1.1 200 OK\n\n Hello, world!\n"'; done`,
//   				},
//   			},
//   		}, nil
//   	}
//
//   	packit.Run(detect, build)
//   }
//
// Summary
//
// These examples show the very basics of what a buildpack implementation using
// packit might entail. For more details, please consult the documentation of
// the types and functions declared herein.
package packit
