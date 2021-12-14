// Package draft provides a service for resolving the priority of buildpack
// plan entries as well as consilidating build and launch requirements.
package draft

import (
	"reflect"
	"regexp"
	"sort"

	"github.com/buildpacks/libcnb"
)

// A Planner sorts buildpack plan entries using a given list of priorities. A
// Planner can also give the OR merged state of launch and build fields that
// are defined in the buildpack plan entries metadata field.
type Planner struct {
}

// NewPlanner returns a new Planner object.
func NewPlanner() Planner {
	return Planner{}
}

// Resolve takes the name of buildpack plan entries that you want to sort, the
// buildpack plan entries that you want to be sorted, and a priority list of
// version-sources where the 0th index is the highest priority. Priorities can
// either be a string, in which case an exact string match with the
// version-source wil be required, or it can be a regular expression. It
// returns the highest priority entry as well as the sorted and filtered list
// of buildpack plan entries that were given. Entries with no given
// version-source are the lowest priority.
//
// If nil is passed for the value of the priority list then the function will
// just return the first filtered entry from the list of the entries that were
// passed into the function initially.
func (p Planner) Resolve(name string, entries []libcnb.BuildpackPlanEntry, priorities []interface{}) (libcnb.BuildpackPlanEntry, []libcnb.BuildpackPlanEntry) {
	var filteredEntries []libcnb.BuildpackPlanEntry
	for _, e := range entries {
		if e.Name == name {
			filteredEntries = append(filteredEntries, e)
		}
	}

	if len(filteredEntries) == 0 {
		return libcnb.BuildpackPlanEntry{}, nil
	}

	sort.Slice(filteredEntries, func(i, j int) bool {
		leftSource := filteredEntries[i].Metadata["version-source"]
		left, _ := leftSource.(string)
		leftPriority := -1

		rightSource := filteredEntries[j].Metadata["version-source"]
		right, _ := rightSource.(string)
		rightPriority := -1

		for index, match := range priorities {
			if r, ok := match.(*regexp.Regexp); ok {
				if r.MatchString(left) {
					leftPriority = len(priorities) - index - 1
				}
			} else {
				if reflect.DeepEqual(match, left) {
					leftPriority = len(priorities) - index - 1
				}
			}

			if r, ok := match.(*regexp.Regexp); ok {
				if r.MatchString(right) {
					rightPriority = len(priorities) - index - 1
				}
			} else {
				if reflect.DeepEqual(match, right) {
					rightPriority = len(priorities) - index - 1
				}
			}
		}

		return leftPriority > rightPriority
	})

	return filteredEntries[0], filteredEntries
}

// MergeLayerTypes takes the name of buildpack plan entries that you want and
// the list buildpack plan entries you want merged layered types from. It
// returns the OR result of the launch and build keys for all of the buildpack
// plan entries with the specified name. The first return is the value of the
// OR launch the second return value is OR build.
func (p Planner) MergeLayerTypes(name string, entries []libcnb.BuildpackPlanEntry) (bool, bool) {
	var launch, build bool
	for _, e := range entries {
		if e.Name == name {
			for _, phase := range []string{"build", "launch"} {
				if e.Metadata[phase] == true {
					switch phase {
					case "build":
						build = true
					case "launch":
						launch = true
					}
				}
			}
		}
	}

	return launch, build
}
