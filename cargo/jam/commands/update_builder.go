package commands

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

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
		panic(err)
	}

	if options.BuilderFile == "" {
		panic("no builder-file flag")
		// return errors.New("--builder-file is a required flag")
	}

	builder, err := internal.ParseBuilderConfig(options.BuilderFile)
	if err != nil {
		panic(err)
	}

	for i, buildpack := range builder.Buildpacks {
		image, err := internal.FindLatestImage(buildpack.URI)
		if err != nil {
			panic(err)
			// return err
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

	uri, err := url.Parse(options.LifecycleURI)
	if err != nil {
		panic(err)
		// return err
	}

	uri.Scheme = ""

	lifecycleURI := strings.TrimPrefix(uri.String(), "//")

	image, err := internal.FindLatestImage(lifecycleURI)
	if err != nil {
		panic(err)
		// return err
	}

	builder.Lifecycle.Version = image.Version

	err = internal.OverwriteBuilderConfig(options.BuilderFile, builder)
	if err != nil {
		panic(err)
	}

	return nil
}
