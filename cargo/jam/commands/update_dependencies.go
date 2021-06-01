package commands

import (
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/spf13/cobra"
)

type updateDependenciesFlags struct {
	buildpackFile string
	api           string
}

func updateDependencies() *cobra.Command {
	flags := &updateDependenciesFlags{}
	cmd := &cobra.Command{
		Use:   "update-dependencies",
		Short: "updates all depdendencies in a buildpack.toml according to metadata.constraints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateDependenciesRun(*flags)
		},
	}
	cmd.Flags().StringVar(&flags.buildpackFile, "buildpack-file", "", "path to the buildpack.toml file (required)")
	cmd.Flags().StringVar(&flags.api, "api", "https://api.deps.paketo.io", "api to query for dependencies")

	err := cmd.MarkFlagRequired("buildpack-file")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to mark buildpack-file flag as required")
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(updateDependencies())
}

func updateDependenciesRun(flags updateDependenciesFlags) error {
	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(flags.buildpackFile)
	if err != nil {
		return fmt.Errorf("failed to parse buildpack.toml: %s", err)
	}

	api := flags.api

	// All internal.Dependencies from the dep-server
	allDependencies := map[string][]internal.Dependency{}
	// All cargo.ConfigMetadataDependencies that match one of the given constraints
	var matchingDependencies []cargo.ConfigMetadataDependency

	for _, constraint := range config.Metadata.DependencyConstraints {
		// Only query the API once per unique dependency
		dependencies, ok := allDependencies[constraint.ID]
		if !ok {
			var err error
			fmt.Printf("reaching out to %s/v1/dependency?name=%s", api, constraint.ID)
			dependencies, err = internal.GetAllDependencies(api, constraint.ID)
			if err != nil {
				return err
			}
			allDependencies[constraint.ID] = dependencies
		}

		// Manually lookup the existent dependency name from the buildpack.toml since this
		// isn't specified via the dep-server
		dependencyName := internal.FindDependencyName(constraint.ID, config)

		// Filter allDependencies for only those that match the constraint
		mds, err := internal.GetDependenciesWithinConstraint(dependencies, constraint, dependencyName)
		if err != nil {
			return err
		}
		matchingDependencies = append(matchingDependencies, mds...)
	}

	if len(matchingDependencies) > 0 {
		config.Metadata.Dependencies = matchingDependencies
	}

	file, err := os.OpenFile(flags.buildpackFile, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open buildpack config file: %w", err)
	}
	defer file.Close()

	err = cargo.EncodeConfig(file, config)
	if err != nil {
		return fmt.Errorf("failed to write buildpack config: %w", err)
	}

	return nil
}
