package scribe

import (
	"io"
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
