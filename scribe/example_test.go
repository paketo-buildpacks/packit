package scribe_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func ExampleEmitter() {
	emitter := scribe.NewEmitter(os.Stdout)

	emitter.Title("Title")
	emitter.Process("Process")
	emitter.Subprocess("Subprocess")
	emitter.Action("Action")
	emitter.Detail("Detail")
	emitter.Subdetail("Subdetail")
	emitter.Break()
	emitter.Title("Next line")

	// Output:
	// Title
	//   Process
	//     Subprocess
	//       Action
	//         Detail
	//           Subdetail
	//
	// Next line
}

func ExampleEmitter_SelectedDependency() {
	emitter := scribe.NewEmitter(os.Stdout)

	deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
	if err != nil {
		log.Fatal(err)
	}

	entry := libcnb.BuildpackPlanEntry{
		Metadata: map[string]interface{}{"version-source": "some-source"},
	}

	dependency := postal.Dependency{
		Name:            "Some Dependency",
		Version:         "some-version",
		DeprecationDate: deprecationDate,
	}

	emitter.Title("SelectedDependency")
	emitter.SelectedDependency(entry, dependency, deprecationDate.Add(-30*24*time.Hour))
	emitter.SelectedDependency(entry, dependency, deprecationDate.Add(-29*24*time.Hour))
	emitter.SelectedDependency(entry, dependency, deprecationDate.Add(24*time.Hour))

	// Output:
	// SelectedDependency
	//     Selected Some Dependency version (using some-source): some-version
	//
	//     Selected Some Dependency version (using some-source): some-version
	//       Version some-version of Some Dependency will be deprecated after 2021-04-01.
	//       Migrate your application to a supported version of Some Dependency before this time.
	//
	//     Selected Some Dependency version (using some-source): some-version
	//       Version some-version of Some Dependency is deprecated.
	//       Migrate your application to a supported version of Some Dependency.
	//
}

func ExampleEmitter_Candidates() {
	emitter := scribe.NewEmitter(os.Stdout)

	emitter.Candidates([]libcnb.BuildpackPlanEntry{
		{
			Metadata: map[string]interface{}{
				"version-source": "some-source",
				"version":        "some-version",
			},
		},
		{
			Metadata: map[string]interface{}{
				"version": "other-version",
			},
		},
	})

	// Output:
	//     Candidate version sources (in priority order):
	//       some-source -> "some-version"
	//       <unknown>   -> "other-version"
	//
}

func ExampleEmitter_LaunchProcesses() {
	emitter := scribe.NewEmitter(os.Stdout)

	processes := []libcnb.Process{
		{
			Type:    "some-type",
			Command: "some-command",
		},
		{
			Type:    "web",
			Command: "web-command",
			Default: true,
		},
		{
			Type:      "some-other-type",
			Command:   "some-other-command",
			Arguments: []string{"some", "args"},
		},
	}

	processEnvs := []map[string]libcnb.Environment{
		{
			"web": libcnb.Environment{
				"WEB_VAR.default": "some-env",
			},
		},
		{
			"web": libcnb.Environment{
				"ANOTHER_WEB_VAR.default": "another-env",
			},
		},
	}

	emitter.LaunchProcesses(processes)
	emitter.LaunchProcesses(processes, processEnvs...)

	// Output:
	//   Assigning launch processes:
	//     some-type:       some-command
	//     web (default):   web-command
	//     some-other-type: some-other-command some args

	//   Assigning launch processes:
	//     some-type:       some-command
	//     web (default):   web-command
	//       ANOTHER_WEB_VAR -> "another-env"
	//       WEB_VAR         -> "some-env"
	//     some-other-type: some-other-command some args
}

func ExampleEmitter_EnvironmentVariables() {
	emitter := scribe.NewEmitter(os.Stdout)

	emitter.EnvironmentVariables(libcnb.Layer{
		BuildEnvironment: libcnb.Environment{
			"NODE_HOME.default":    "/some/path",
			"NODE_ENV.default":     "some-env",
			"NODE_VERBOSE.default": "some-bool",
		},
		LaunchEnvironment: libcnb.Environment{
			"NODE_HOME.default":    "/some/path",
			"NODE_ENV.default":     "another-env",
			"NODE_VERBOSE.default": "another-bool",
		},
		SharedEnvironment: libcnb.Environment{
			"SHARED_ENV.default": "shared-env",
		},
	})

	// Output:
	//   Configuring build environment
	//     NODE_ENV     -> "some-env"
	//     NODE_HOME    -> "/some/path"
	//     NODE_VERBOSE -> "some-bool"
	//     SHARED_ENV   -> "shared-env"
	//
	//   Configuring launch environment
	//     NODE_ENV     -> "another-env"
	//     NODE_HOME    -> "/some/path"
	//     NODE_VERBOSE -> "another-bool"
	//     SHARED_ENV   -> "shared-env"
	//
}

func ExampleFormattedList() {
	fmt.Println(scribe.FormattedList{
		"third",
		"first",
		"second",
	}.String())

	// Output:
	// first
	// second
	// third
}

func ExampleFormattedMap() {
	fmt.Println(scribe.FormattedMap{
		"third":  3,
		"first":  1,
		"second": 2,
	}.String())

	// Output:
	// first  -> "1"
	// second -> "2"
	// third  -> "3"
}

func ExampleNewFormattedMapFromEnvironment() {
	fmt.Println(scribe.NewFormattedMapFromEnvironment(libcnb.Environment{
		"OVERRIDE.override": "some-value",
		"DEFAULT.default":   "some-value",
		"PREPEND.prepend":   "some-value",
		"PREPEND.delim":     ":",
		"APPEND.append":     "some-value",
		"APPEND.delim":      ":",
	}).String())

	// Output:
	// APPEND   -> "$APPEND:some-value"
	// DEFAULT  -> "some-value"
	// OVERRIDE -> "some-value"
	// PREPEND  -> "some-value:$PREPEND"
}

func ExampleLogger() {
	logger := scribe.NewLogger(os.Stdout)

	logger.Title("Title")
	logger.Process("Process")
	logger.Subprocess("Subprocess")
	logger.Action("Action")
	logger.Detail("Detail")
	logger.Subdetail("Subdetail")
	logger.Break()
	logger.Title("Next line")

	// Output:
	// Title
	//   Process
	//     Subprocess
	//       Action
	//         Detail
	//           Subdetail
	//
	// Next line
}

func ExampleLogger_WithLevel() {
	logger := scribe.NewLogger(os.Stdout)

	logger.Title("First line")
	logger.Debug.Title("Debug line")
	logger.Title("Next line")
	logger.Break()

	logger = logger.WithLevel("DEBUG")

	logger.Title("First line")
	logger.Debug.Title("Debug line")
	logger.Title("Next line")

	// Output:
	// First line
	// Next line
	//
	// First line
	// Debug line
	// Next line
}
