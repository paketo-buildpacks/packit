package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/Masterminds/semver/v3"
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

type depKey [3]string

func printImplementation(writer io.Writer, config cargo.Config) {
	if len(config.Stacks) > 0 {
		sort.Slice(config.Stacks, func(i, j int) bool {
			return config.Stacks[i].ID < config.Stacks[j].ID
		})

		fmt.Fprintf(writer, "#### Supported Stacks:\n")
		for _, s := range config.Stacks {
			fmt.Fprintf(writer, "- %s\n", s.ID)
		}
		fmt.Fprintln(writer)
	}

	if len(config.Metadata.DefaultVersions) > 0 {
		fmt.Fprintf(writer, "#### Default Dependency Versions:\n| ID | Version |\n|---|---|\n")
		var sortedDependencies []string
		for key := range config.Metadata.DefaultVersions {
			sortedDependencies = append(sortedDependencies, key)
		}

		sort.Strings(sortedDependencies)

		for _, key := range sortedDependencies {
			fmt.Fprintf(writer, "| %s | %s |\n", key, config.Metadata.DefaultVersions[key])
		}
		fmt.Fprintln(writer)
	}

	if len(config.Metadata.Dependencies) > 0 {
		infoMap := map[depKey][]string{}
		for _, d := range config.Metadata.Dependencies {
			key := depKey{d.ID, d.Version, d.SHA256}
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
				SHA256:  key[2],
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

		fmt.Fprintf(writer, "#### Dependencies:\n| Name | Version | SHA256 |\n|---|---|---|\n")
		for _, d := range sorted {
			fmt.Fprintf(writer, "| %s | %s | %s |\n", d.ID, d.Version, d.SHA256)
		}
		fmt.Fprintln(writer)
	}

}

func (f Formatter) Markdown(configs []cargo.Config) {
	//Language-family case
	if len(configs) > 1 {
		var familyConfig cargo.Config
		for index, config := range configs {
			if len(config.Order) > 0 {
				familyConfig = config
				configs = append(configs[:index], configs[index+1:]...)
				break
			}
		}

		//Header section
		fmt.Fprintf(f.writer, "## %s %s\n\n**ID:** `%s`\n\n", familyConfig.Buildpack.Name, familyConfig.Buildpack.Version, familyConfig.Buildpack.ID)
		fmt.Fprintf(f.writer, "**Digest:** `%s`\n\n", familyConfig.Buildpack.SHA256)
		fmt.Fprintf(f.writer, "#### Included Buildpackages:\n")
		fmt.Fprintf(f.writer, "| Name | ID | Version |\n|---|---|---|\n")
		for _, config := range configs {
			fmt.Fprintf(f.writer, "| %s | %s | %s |\n", config.Buildpack.Name, config.Buildpack.ID, config.Buildpack.Version)
		}
		//Sub Header
		fmt.Fprintf(f.writer, "\n<details>\n<summary>Order Groupings</summary>\n\n")
		for _, o := range familyConfig.Order {
			fmt.Fprintf(f.writer, "| ID | Version | Optional |\n|---|---|---|\n")
			for _, g := range o.Group {
				fmt.Fprintf(f.writer, "| %s | %s | %t |\n", g.ID, g.Version, g.Optional)
			}
			fmt.Fprintln(f.writer)
		}
		fmt.Fprintf(f.writer, "</details>\n\n---\n")

		for _, config := range configs {
			fmt.Fprintf(f.writer, "\n<details>\n<summary>%s %s</summary>\n", config.Buildpack.Name, config.Buildpack.Version)
			fmt.Fprintf(f.writer, "\n**ID:** `%s`\n\n", config.Buildpack.ID)
			printImplementation(f.writer, config)
			fmt.Fprintf(f.writer, "---\n\n</details>\n")
		}

	} else { //Implementation case
		fmt.Fprintf(f.writer, "## %s %s\n", configs[0].Buildpack.Name, configs[0].Buildpack.Version)
		fmt.Fprintf(f.writer, "\n**ID:** `%s`\n\n", configs[0].Buildpack.ID)
		fmt.Fprintf(f.writer, "**Digest:** `%s`\n\n", configs[0].Buildpack.SHA256)
		printImplementation(f.writer, configs[0])
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
