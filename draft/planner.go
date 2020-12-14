package draft

import (
	"sort"

	"github.com/paketo-buildpacks/packit"
)

type Planner struct {
}

func NewPlanner() Planner {
	return Planner{}
}

func (p Planner) Resolve(name string, entries []packit.BuildpackPlanEntry, priorities map[string]int) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry) {
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

		rightSource := filteredEntries[j].Metadata["version-source"]
		right, _ := rightSource.(string)

		return priorities[left] > priorities[right]
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
