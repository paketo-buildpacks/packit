package internal

import (
	"encoding/json"
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

func lookupName(configs []cargo.Config, id string) string {
	for _, config := range configs {
		if config.Buildpack.ID == id {
			return config.Buildpack.Name
		}
	}

	return ""
}

func (f Formatter) Markdown(configs []cargo.Config) {
	var familyConfig cargo.Config
	for index, config := range configs {
		if len(config.Order) > 0 {
			familyConfig = config
			configs = append(configs[:index], configs[index+1:]...)
			break
		}
	}

	if (familyConfig.Buildpack != cargo.ConfigBuildpack{}) {
		fmt.Fprintf(f.writer, "# %s %s\n**ID:** %s\n\n", familyConfig.Buildpack.Name, familyConfig.Buildpack.Version, familyConfig.Buildpack.ID)
		fmt.Fprintf(f.writer, "| Name | ID | Version |\n|---|---|---|\n")
		for _, config := range configs {
			fmt.Fprintf(f.writer, "| %s | %s | %s |\n", config.Buildpack.Name, config.Buildpack.ID, config.Buildpack.Version)
		}
		fmt.Fprintf(f.writer, "\n<details>\n<summary>More Information</summary>\n\n")
		fmt.Fprintf(f.writer, "### Order Groupings\n")
		for _, o := range familyConfig.Order {
			fmt.Fprintf(f.writer, "| Name | ID | Version | Optional |\n|---|---|---|---|\n")
			for _, g := range o.Group {
				fmt.Fprintf(f.writer, "| %s | %s | %s | %t |\n", lookupName(configs, g.ID), g.ID, g.Version, g.Optional)
			}
			fmt.Fprintln(f.writer)
		}
	}

	for _, config := range configs {
		fmt.Fprintf(f.writer, "## %s %s\n**ID:** %s\n\n", config.Buildpack.Name, config.Buildpack.Version, config.Buildpack.ID)

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

			fmt.Fprintf(f.writer, "### Dependencies\n| Name | Version | Stacks |\n|---|---|---|\n")
			for _, d := range sorted {
				fmt.Fprintf(f.writer, "| %s | %s | %s |\n", d.ID, d.Version, strings.Join(d.Stacks, ", "))
			}
			fmt.Fprintln(f.writer)
		}

		if len(config.Metadata.DefaultVersions) > 0 {
			fmt.Fprintf(f.writer, "### Default Dependencies\n| Name | Version |\n|---|---|\n")
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

			fmt.Fprintf(f.writer, "### Supported Stacks\n| Name |\n|---|\n")
			for _, s := range config.Stacks {
				fmt.Fprintf(f.writer, "| %s |\n", s.ID)
			}
			fmt.Fprintln(f.writer)
		}
	}

	if (familyConfig.Buildpack != cargo.ConfigBuildpack{}) {
		fmt.Fprintln(f.writer, "</details>")
	}
}

func (f Formatter) JSON(configs []cargo.Config) {
	var output struct {
		Buildpackage cargo.Config   `json:"buildpackage"`
		Children     []cargo.Config `json:"children,omitempty"`
	}

	output.Buildpackage = configs[0]

	if len(configs) > 1 {
		for _, config := range configs {
			if len(config.Order) > 0 {
				output.Buildpackage = config
			} else {
				output.Children = append(output.Children, config)
			}
		}
	}

	_ = json.NewEncoder(f.writer).Encode(&output)
}
