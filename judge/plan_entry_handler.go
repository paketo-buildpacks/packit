package judge

import (
	"sort"
	"strconv"

	"github.com/paketo-buildpacks/packit"
)

type Logger interface {
	Subprocess(string, ...interface{})
	Action(string, ...interface{})
	Break()
}

type Emitter interface {
	Candidates([]packit.BuildpackPlanEntry)
}

type PlanEntryHandler struct {
	logger Logger
}

func NewPlanEntryHandler(logger Logger) PlanEntryHandler {
	return PlanEntryHandler{
		logger: logger,
	}
}

func (p PlanEntryHandler) ResolveEntries(name string, entries []packit.BuildpackPlanEntry, priorities map[string]int) (packit.BuildpackPlanEntry, bool) {
	var filteredEntries []packit.BuildpackPlanEntry
	for _, e := range entries {
		if e.Name == name {
			filteredEntries = append(filteredEntries, e)
		}
	}

	if len(filteredEntries) == 0 {
		return packit.BuildpackPlanEntry{}, false
	}

	sort.Slice(filteredEntries, func(i, j int) bool {
		leftSource := filteredEntries[i].Metadata["version-source"]
		left, _ := leftSource.(string)

		rightSource := filteredEntries[j].Metadata["version-source"]
		right, _ := rightSource.(string)

		return priorities[left] > priorities[right]
	})

	emitter, ok := p.logger.(Emitter)
	if ok {
		emitter.Candidates(filteredEntries)
	} else {
		p.candidates(filteredEntries)
	}

	return filteredEntries[0], true
}

func (p PlanEntryHandler) MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) []packit.LayerType {
	layerTypeCollector := map[packit.LayerType]interface{}{}
	for _, e := range entries {
		if e.Name == name {
			for _, phase := range []string{"build", "launch"} {
				if e.Metadata[phase] == true {
					switch phase {
					case "build":
						layerTypeCollector[packit.BuildLayer] = nil
						layerTypeCollector[packit.CacheLayer] = nil
					case "launch":
						layerTypeCollector[packit.LaunchLayer] = nil
					}
				}
			}
		}
	}

	var layerTypes []packit.LayerType
	for layerType := range layerTypeCollector {
		layerTypes = append(layerTypes, layerType)
	}

	return layerTypes
}

func (p PlanEntryHandler) candidates(entries []packit.BuildpackPlanEntry) {
	p.logger.Subprocess("Candidate version sources (in priority order):")

	var (
		sources [][2]string
		maxLen  int
	)

	for _, entry := range entries {
		versionSource, ok := entry.Metadata["version-source"].(string)
		if !ok {
			versionSource = "<unknown>"
		}

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = "*"
		}

		if len(versionSource) > maxLen {
			maxLen = len(versionSource)
		}

		sources = append(sources, [2]string{versionSource, version})
	}

	for _, source := range sources {
		p.logger.Action(("%-" + strconv.Itoa(maxLen) + "s -> %q"), source[0], source[1])
	}

	p.logger.Break()
}
