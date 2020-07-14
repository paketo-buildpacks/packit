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

func (f Formatter) Markdown(config cargo.Config) {
	titlePrefix := "##"
	if len(config.Order) > 0 {
		titlePrefix = "#"
	}

	fmt.Fprintf(f.writer, "%s %s %s\n", titlePrefix, config.Buildpack.ID, config.Buildpack.Version)

	if len(config.Order) > 0 {
		fmt.Fprintf(f.writer, "### Order Groupings\n")
		for _, o := range config.Order {
			fmt.Fprintf(f.writer, "| name | version | optional |\n|-|-|-|\n")
			for _, g := range o.Group {
				fmt.Fprintf(f.writer, "| %s | %s | %t |\n", g.ID, g.Version, g.Optional)
			}
			fmt.Fprintln(f.writer)
		}
	}

	if len(config.Metadata.Dependencies) > 0 {
		infoMap := map[depKey][]string{}
		for _, d := range config.Metadata.Dependencies {
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

		fmt.Fprintf(f.writer, "### Dependencies\n| name | version | stacks |\n|-|-|-|\n")
		for _, d := range sorted {
			fmt.Fprintf(f.writer, "| %s | %s | %s |\n", d.ID, d.Version, strings.Join(d.Stacks, ", "))
		}
		fmt.Fprintln(f.writer)
	}

	if len(config.Metadata.DefaultVersions) > 0 {
		fmt.Fprintf(f.writer, "### Default Dependencies\n| name | version |\n|-|-|\n")
		var sortedDependencies []string
		for key := range config.Metadata.DefaultVersions {
			sortedDependencies = append(sortedDependencies, key)
		}

		sort.Strings(sortedDependencies)

		for _, key := range sortedDependencies {
			fmt.Fprintf(f.writer, "| %s | %s |\n", key, config.Metadata.DefaultVersions[key])
		}
		fmt.Fprintln(f.writer)
	}

	if len(config.Stacks) > 0 {
		sort.Slice(config.Stacks, func(i, j int) bool {
			return config.Stacks[i].ID < config.Stacks[j].ID
		})

		fmt.Fprintf(f.writer, "### Supported Stacks\n| name |\n|-|\n")
		for _, s := range config.Stacks {
			fmt.Fprintf(f.writer, "| %s |\n", s.ID)
		}
	}
}
