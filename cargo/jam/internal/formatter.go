package internal

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/cargo"
)

type Formatter struct {
	writer io.Writer
}

func NewFormatter(writer io.Writer) Formatter {
	return Formatter{
		writer: writer,
	}
}

type depKey [2]string

func (f Formatter) Markdown(dependencies []cargo.ConfigMetadataDependency, defaults map[string]string, stacks []string) {
	if len(dependencies) > 0 {
		infoMap := map[depKey][]string{}
		for _, d := range dependencies {
			key := depKey{d.ID, d.Version}
			_, ok := infoMap[key]
			if !ok {
				sort.Strings(d.Stacks)
				infoMap[key] = d.Stacks
			} else {
				val := infoMap[key]
				val = append(val, d.Stacks...)
				sort.Strings(val)
				infoMap[key] = val
			}
		}

		var sorted []cargo.ConfigMetadataDependency
		for key, stacks := range infoMap {
			sorted = append(sorted, cargo.ConfigMetadataDependency{
				ID:      key[0],
				Version: key[1],
				Stacks:  stacks,
			})
		}

		sort.Slice(sorted, func(i, j int) bool {
			iVal := sorted[i]
			jVal := sorted[j]

			if iVal.ID < jVal.ID {
				return true
			}

			iVersion := semver.MustParse(iVal.Version)
			jVersion := semver.MustParse(jVal.Version)

			return iVal.ID == jVal.ID && iVersion.GreaterThan(jVersion)
		})

		fmt.Fprintf(f.writer, "Dependencies:\n| name | version | stacks |\n|-|-|-|\n")
		for _, d := range sorted {
			fmt.Fprintf(f.writer, "| %s | %s | %s |\n", d.ID, d.Version, strings.Join(d.Stacks, ", "))
		}
		fmt.Fprintln(f.writer)
	}

	if len(defaults) > 0 {
		fmt.Fprintf(f.writer, "Default dependencies:\n| name | version |\n|-|-|\n")
		var sortedDependencies []string
		for key := range defaults {
			sortedDependencies = append(sortedDependencies, key)
		}

		sort.Strings(sortedDependencies)

		for _, key := range sortedDependencies {
			fmt.Fprintf(f.writer, "| %s | %s |\n", key, defaults[key])
		}
		fmt.Fprintln(f.writer)
	}

	if len(stacks) > 0 {
		sort.Strings(stacks)

		fmt.Fprintf(f.writer, "Supported stacks:\n| name |\n|-|\n")
		for _, s := range stacks {
			fmt.Fprintf(f.writer, "| %s |\n", s)
		}
	}
}
