package draft_test

import (
	"fmt"
	"regexp"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/packit/v2/draft"
)

func ExamplePlanner_Resolve() {
	buildpackPlanEntries := []libcnb.BuildpackPlanEntry{
		{
			Name: "fred",
		},
		{
			Name: "clint",
			Metadata: map[string]interface{}{
				"version-source": "high",
			},
		},
		{
			Name: "fred",
			Metadata: map[string]interface{}{
				"version-source": "high",
			},
		},
		{
			Name: "fred",
			Metadata: map[string]interface{}{
				"version-source": "some-low-priority",
			},
		},
	}

	priorities := []interface{}{"high", regexp.MustCompile(`.*low.*`)}

	planner := draft.NewPlanner()

	entry, entries := planner.Resolve("fred", buildpackPlanEntries, priorities)

	printEntry := func(e libcnb.BuildpackPlanEntry) {
		var source string
		source, ok := e.Metadata["version-source"].(string)
		if !ok {
			source = ""
		}
		fmt.Printf("%s => %q\n", e.Name, source)
	}

	fmt.Println("Highest Priority Entry")
	printEntry(entry)
	fmt.Println("Buildpack Plan Entry List Priority Sorted")
	for _, e := range entries {
		printEntry(e)
	}

	// Output:
	// Highest Priority Entry
	// fred => "high"
	// Buildpack Plan Entry List Priority Sorted
	// fred => "high"
	// fred => "some-low-priority"
	// fred => ""
}

func ExamplePlanner_MergeLayerTypes() {
	buildpackPlanEntries := []libcnb.BuildpackPlanEntry{
		{
			Name: "fred",
		},
		{
			Name: "clint",
			Metadata: map[string]interface{}{
				"build": false,
			},
		},
		{
			Name: "fred",
			Metadata: map[string]interface{}{
				"build": true,
			},
		},
		{
			Name: "fred",
			Metadata: map[string]interface{}{
				"launch": true,
			},
		},
	}

	planner := draft.NewPlanner()

	launch, build := planner.MergeLayerTypes("fred", buildpackPlanEntries)

	fmt.Printf("launch => %t; build => %t", launch, build)

	// Output:
	// launch => true; build => true
}
