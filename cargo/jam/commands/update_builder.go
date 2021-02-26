package commands

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/spf13/cobra"
)

type updateBuilderFlags struct {
	builderFile  string
	lifecycleURI string
}

func updateBuilder() *cobra.Command {
	flags :=&updateBuilderFlags{}
	cmd := &cobra.Command{
		Use:   "update-builder",
		Short: "update builder",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateBuilderRun(*flags)
		},
	}
	cmd.Flags().StringVar(&flags.builderFile, "builder-file", "", "path to the builder.toml file (required)")
	cmd.Flags().StringVar(&flags.lifecycleURI, "lifecycle-uri", "index.docker.io/buildpacksio/lifecycle", "URI for lifecycle image (optional: default=index.docker.io/buildpacksio/lifecycle)")

	cmd.MarkFlagRequired("builder-file")
	return cmd
}

func init() {
	rootCmd.AddCommand(updateBuilder())
}

func updateBuilderRun(flags updateBuilderFlags) error {
	builder, err := internal.ParseBuilderConfig(flags.builderFile)
	if err != nil {
		return err
	}

	for i, buildpack := range builder.Buildpacks {
		image, err := internal.FindLatestImage(buildpack.URI)
		if err != nil {
			return err
		}

		builder.Buildpacks[i].Version = image.Version
		builder.Buildpacks[i].URI = fmt.Sprintf("%s:%s", image.Name, image.Version)

		for j, order := range builder.Order {
			for k, group := range order.Group {
				if group.ID == image.Path {
					builder.Order[j].Group[k].Version = image.Version
				}
			}
		}
	}

	lifecycleImage, err := internal.FindLatestImage(flags.lifecycleURI)
	if err != nil {
		return err
	}

	builder.Lifecycle.Version = lifecycleImage.Version

	buildImage, err := internal.FindLatestBuildImage(builder.Stack.RunImage, builder.Stack.BuildImage)
	if err != nil {
		return err
	}

	builder.Stack.BuildImage = fmt.Sprintf("%s:%s", buildImage.Name, buildImage.Version)

	err = internal.OverwriteBuilderConfig(flags.builderFile, builder)
	if err != nil {
		return err
	}

	return nil
}
