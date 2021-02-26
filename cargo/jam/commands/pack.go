package commands

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type packFlags struct {
	buildpackTOMLPath string
	output            string
	version           string
	offline           bool
	stack             string
}

func pack() *cobra.Command {
	flags :=&packFlags{}
	cmd := &cobra.Command{
		Use:   "pack",
		Short: "package buildpack",
		RunE: func(cmd *cobra.Command, args []string) error {
			return packRun(*flags)
		},
	}
	cmd.Flags().StringVar(&flags.buildpackTOMLPath, "buildpack", "", "path to buildpack.toml")
	cmd.Flags().StringVar(&flags.output, "output", "", "path to location of output tarball")
	cmd.Flags().StringVar(&flags.version, "version", "", "version of the buildpack")
	cmd.Flags().BoolVar(&flags.offline, "offline", false, "enable offline caching of dependencies")
	cmd.Flags().StringVar(&flags.stack, "stack", "", "restricts dependencies to given stack")

	err := cmd.MarkFlagRequired("buildpack")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to mark buildpack flag as required")
	}
	err = cmd.MarkFlagRequired("version")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to mark version flag as required")
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(pack())
}

func packRun(flags packFlags) error {
	buildpackDir, err := ioutil.TempDir("", "dup-dest")
	if err != nil {
		return fmt.Errorf("unable to create temporary directory: %s", err)
	}
	defer os.RemoveAll(buildpackDir)

	directoryDuplicator := cargo.NewDirectoryDuplicator()
	err = directoryDuplicator.Duplicate(filepath.Dir(flags.buildpackTOMLPath), buildpackDir)
	if err != nil {
		return fmt.Errorf("failed to duplicate directory: %s", err)
	}

	flags.buildpackTOMLPath = filepath.Join(buildpackDir, filepath.Base(flags.buildpackTOMLPath))

	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(flags.buildpackTOMLPath)
	if err != nil {
		return fmt.Errorf("failed to parse buildpack.toml: %s", err)
	}

	config.Buildpack.Version = flags.version

	fmt.Fprintf(os.Stdout, "Packing %s %s...\n", config.Buildpack.Name, flags.version)

	if flags.stack != "" {
		var filteredDependencies []cargo.ConfigMetadataDependency
		for _, dep := range config.Metadata.Dependencies {
			if dep.HasStack(flags.stack) {
				filteredDependencies = append(filteredDependencies, dep)
			}
		}

		config.Metadata.Dependencies = filteredDependencies
	}

	logger := scribe.NewLogger(os.Stdout)
	bash := pexec.NewExecutable("bash")
	prePackager := internal.NewPrePackager(bash, logger, scribe.NewWriter(os.Stdout, scribe.WithIndent(2)))
	err = prePackager.Execute(config.Metadata.PrePackage, buildpackDir)
	if err != nil {
		return fmt.Errorf("failed to execute pre-packaging script %q: %s", config.Metadata.PrePackage, err)
	}

	if flags.offline {
		transport := cargo.NewTransport()
		dependencyCacher := internal.NewDependencyCacher(transport, logger)
		config.Metadata.Dependencies, err = dependencyCacher.Cache(buildpackDir, config.Metadata.Dependencies)
		if err != nil {
			return fmt.Errorf("failed to cache dependencies: %s", err)
		}

		for _, dependency := range config.Metadata.Dependencies {
			config.Metadata.IncludeFiles = append(config.Metadata.IncludeFiles, strings.TrimPrefix(dependency.URI, "file:///"))
		}
	}

	fileBundler := internal.NewFileBundler()
	files, err := fileBundler.Bundle(buildpackDir, config.Metadata.IncludeFiles, config)
	if err != nil {
		return fmt.Errorf("failed to bundle files: %s", err)
	}

	tarBuilder := internal.NewTarBuilder(logger)
	err = tarBuilder.Build(flags.output, files)
	if err != nil {
		return fmt.Errorf("failed to create output: %s", err)
	}

	return nil
}
