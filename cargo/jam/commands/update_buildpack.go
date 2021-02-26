package commands

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/spf13/cobra"
)

type updateBuildpackFlags struct {
	buildpackFile  string
	packageFile string
}

func updateBuildpack() *cobra.Command {
	flags :=&updateBuildpackFlags{}
	cmd := &cobra.Command{
		Use:   "update-buildpack",
		Short: "update buildpack",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateBuildpackRun(*flags)
		},
	}
	cmd.Flags().StringVar(&flags.buildpackFile, "buildpack-file", "", "path to the buildpack.toml file (required)")
	cmd.Flags().StringVar(&flags.packageFile, "package-file", "", "path to the package.toml file (required)")

	cmd.MarkFlagRequired("buildpack-file")
	cmd.MarkFlagRequired("package-file")
	return cmd
}

func init() {
	rootCmd.AddCommand(updateBuildpack())
}

func updateBuildpackRun(flags updateBuildpackFlags) error {
	bp, err := internal.ParseBuildpackConfig(flags.buildpackFile)
	if err != nil {
		return err
	}

	pkg, err := internal.ParsePackageConfig(flags.packageFile)
	if err != nil {
		return err
	}

	for i, dependency := range pkg.Dependencies {
		image, err := internal.FindLatestImage(dependency.URI)
		if err != nil {
			return err
		}

		pkg.Dependencies[i].URI = fmt.Sprintf("%s:%s", image.Name, image.Version)

		for j, order := range bp.Order {
			for k, group := range order.Group {
				if group.ID == image.Path {
					bp.Order[j].Group[k].Version = image.Version
				}
			}
		}
	}

	err = internal.OverwriteBuildpackConfig(flags.buildpackFile, bp)
	if err != nil {
		return err
	}

	err = internal.OverwritePackageConfig(flags.packageFile, pkg)
	if err != nil {
		return err
	}

	return nil
}
