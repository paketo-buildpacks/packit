package commands

import (
	"errors"
	"flag"
	"fmt"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
)

type UpdateBuilder struct{}

func NewUpdateBuilder() UpdateBuilder {
	return UpdateBuilder{}
}

func (ub UpdateBuilder) Execute(args []string) error {
	var options struct {
		BuilderFile  string
		LifecycleURI string
	}

	fset := flag.NewFlagSet("update-builder", flag.ContinueOnError)
	fset.StringVar(&options.BuilderFile, "builder-file", "", "path to the builder.toml file (required)")
	fset.StringVar(&options.LifecycleURI, "lifecycle-uri", "index.docker.io/buildpacksio/lifecycle", "URI for lifecycle image (optional: default=index.docker.io/buildpacksio/lifecycle)")
	err := fset.Parse(args)
	if err != nil {
		return err
	}

	if options.BuilderFile == "" {
		return errors.New("--builder-file is a required flag")
	}

	builder, err := internal.ParseBuilderConfig(options.BuilderFile)
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
					if builder.Order[j].Group[k].Version != "" {
						builder.Order[j].Group[k].Version = image.Version
					}
				}
			}
		}
	}

	lifecycleImage, err := internal.FindLatestImage(options.LifecycleURI)
	if err != nil {
		return err
	}

	builder.Lifecycle.Version = lifecycleImage.Version

	buildImage, err := internal.FindLatestBuildImage(builder.Stack.RunImage, builder.Stack.BuildImage)
	if err != nil {
		return err
	}

	builder.Stack.BuildImage = fmt.Sprintf("%s:%s", buildImage.Name, buildImage.Version)

	err = internal.OverwriteBuilderConfig(options.BuilderFile, builder)
	if err != nil {
		return err
	}

	return nil
}
