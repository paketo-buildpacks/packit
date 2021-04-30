package scribe

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

type Emitter struct {
	// Logger is embedded and therefore delegates all of its functions to the
	// Emitter.
	Logger
}

func NewEmitter(output io.Writer) Emitter {
	return Emitter{
		Logger: NewLogger(output),
	}
}

func (e Emitter) SelectedDependency(entry packit.BuildpackPlanEntry, dependency postal.Dependency, now time.Time) {
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

func (e Emitter) Candidates(entries []packit.BuildpackPlanEntry) {
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

func (e Emitter) LaunchProcesses(processes []packit.Process) {
	e.Process("Assigning launch processes:")

	for _, process := range processes {
		p := fmt.Sprintf("%s: %s", process.Type, process.Command)

		if process.Args != nil {
			p = fmt.Sprintf("%s %s", p, strings.Join(process.Args, " "))
		}

		e.Subprocess(p)
	}
	e.Break()
}

func (e Emitter) EnvironmentVariables(layer packit.Layer) {
	buildEnv := packit.Environment{}
	launchEnv := packit.Environment{}

	// Makes deep local copy of the env map on the layer
	for key, value := range layer.BuildEnv {
		buildEnv[key] = value
	}

	for key, value := range layer.LaunchEnv {
		launchEnv[key] = value
	}

	// Merge the shared env map with the launch and build to remove CNB spec
	// specific terminiology from the output
	for key, value := range layer.SharedEnv {
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
