package scribe

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

// An Emitter embeds the scribe.Logger type to provide an interface for
// complicated shared logging tasks.
type Emitter struct {
	// Logger is embedded and therefore delegates all of its functions to the
	// Emitter.
	Logger
}

// NewEmitter returns an emitter that writes to the given output.
func NewEmitter(output io.Writer) Emitter {
	return Emitter{
		Logger: NewLogger(output),
	}
}

// SelectedDependency takes in a buildpack plan entry, a postal dependency, and
// the current time, and prints out a message giving the name and version of
// the dependency as well as the source of the request for that given
// dependency, it will also print a deprecation warning and an EOL warning
// based if the given dependency is set to be deprecated within the next 30 or
// is past that window.
func (e Emitter) SelectedDependency(entry libcnb.BuildpackPlanEntry, dependency postal.Dependency, now time.Time) {
	source, ok := entry.Metadata["version-source"].(string)
	if !ok {
		source = "<unknown>"
	}

	e.Subprocess("Selected %s version (using %s): %s", dependency.Name, source, dependency.Version)

	if (dependency.DeprecationDate != time.Time{}) {
		deprecationDate := dependency.DeprecationDate
		switch {
		case (deprecationDate.Add(-30*24*time.Hour).Before(now) && deprecationDate.After(now)):
			e.Action("Version %s of %s will be deprecated after %s.", dependency.Version, dependency.Name, dependency.DeprecationDate.Format("2006-01-02"))
			e.Action("Migrate your application to a supported version of %s before this time.", dependency.Name)
		case (deprecationDate == now || deprecationDate.Before(now)):
			e.Action("Version %s of %s is deprecated.", dependency.Version, dependency.Name)
			e.Action("Migrate your application to a supported version of %s.", dependency.Name)
		}
	}
	e.Break()
}

// Candidates takes a priority sorted list of buildpack plan entries and prints
// out a formatted table in priority order removing any duplicate entries.
func (e Emitter) Candidates(entries []libcnb.BuildpackPlanEntry) {
	e.Subprocess("Candidate version sources (in priority order):")

	var (
		sources [][2]string
		maxLen  int
	)

Entries:
	for _, entry := range entries {
		versionSource, ok := entry.Metadata["version-source"].(string)
		if !ok {
			versionSource = "<unknown>"
		}

		if len(versionSource) > maxLen {
			maxLen = len(versionSource)
		}

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = ""
		}

		source := [2]string{versionSource, version}

		// Removes any duplicate entries
		for _, s := range sources {
			if s == source {
				continue Entries
			}
		}

		sources = append(sources, source)
	}

	for _, source := range sources {
		e.Action(("%-" + strconv.Itoa(maxLen) + "s -> %q"), source[0], source[1])
	}

	e.Break()
}

// LaunchProcesses take a list of processes and a map of process specific
// enivronment varables and prints out a formatted table including the type
// name, whether or not it is a default process, the command, arguments, and
// any process specific environment variables.
func (e Emitter) LaunchProcesses(processes []libcnb.Process, processEnvs ...map[string]libcnb.Environment) {
	e.Process("Assigning launch processes:")

	var (
		typePadding int
	)

	for _, process := range processes {
		pType := process.Type
		if process.Default {
			pType += " " + "(default)"
		}

		if len(pType) > typePadding {
			typePadding = len(pType)
		}
	}

	for _, process := range processes {
		pType := process.Type
		if process.Default {
			pType += " " + "(default)"
		}

		pad := typePadding + len(process.Command) - len(pType)
		p := fmt.Sprintf("%s: %*s", pType, pad, process.Command)

		if process.Arguments != nil {
			p += " " + strings.Join(process.Arguments, " ")
		}

		e.Subprocess(p)

		// This ensures that the process environment variable is always the same no
		// matter the order of the process envs map list
		processEnv := libcnb.Environment{}
		for _, pEnvs := range processEnvs {
			if env, ok := pEnvs[process.Type]; ok {
				for key, value := range env {
					processEnv[key] = value
				}
			}
		}

		if len(processEnv) != 0 {
			e.Action("%s", NewFormattedMapFromEnvironment(processEnv))
		}

	}
	e.Break()
}

// EnvironmentVariables take a layer and prints out a formatted table of the
// build and launch time environment variables set in the layer.
func (e Emitter) EnvironmentVariables(layer libcnb.Layer) {
	buildEnv := libcnb.Environment{}
	launchEnv := libcnb.Environment{}

	// Makes deep local copy of the env map on the layer
	for key, value := range layer.BuildEnvironment {
		buildEnv[key] = value
	}

	for key, value := range layer.LaunchEnvironment {
		launchEnv[key] = value
	}

	// Merge the shared env map with the launch and build to remove CNB spec
	// specific terminiology from the output
	for key, value := range layer.SharedEnvironment {
		buildEnv[key] = value
		launchEnv[key] = value
	}

	if len(buildEnv) != 0 {
		e.Process("Configuring build environment")
		e.Subprocess("%s", NewFormattedMapFromEnvironment(buildEnv))
		e.Break()
	}

	if len(launchEnv) != 0 {
		e.Process("Configuring launch environment")
		e.Subprocess("%s", NewFormattedMapFromEnvironment(launchEnv))
		e.Break()
	}
}
