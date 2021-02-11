package draft

import (
	"reflect"
	"regexp"
	"sort"

	"github.com/paketo-buildpacks/packit"
)

type Planner struct {
}

func NewPlanner() Planner {
	return Planner{}
}

func (p Planner) Resolve(name string, entries []packit.BuildpackPlanEntry, priorities []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry) {
	var filteredEntries []packit.BuildpackPlanEntry
	for _, e := range entries {
		if e.Name == name {
			filteredEntries = append(filteredEntries, e)
		}
	}

	if len(filteredEntries) == 0 {
		return packit.BuildpackPlanEntry{}, nil
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

func (p Planner) MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (bool, bool) {
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
