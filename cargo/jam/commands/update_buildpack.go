package commands

import (
	"errors"
	"flag"
	"fmt"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
)

type UpdateBuildpack struct{}

func NewUpdateBuildpack() UpdateBuildpack {
	return UpdateBuildpack{}
}

func (ub UpdateBuildpack) Execute(args []string) error {
	var options struct {
		BuildpackFile string
		PackageFile   string
	}

	fset := flag.NewFlagSet("update-buildpack", flag.ContinueOnError)
	fset.StringVar(&options.BuildpackFile, "buildpack-file", "", "path to the buildpack.toml file (required)")
	fset.StringVar(&options.PackageFile, "package-file", "", "path to the package.toml file (required)")
	err := fset.Parse(args)
	if err != nil {
		return err
	}

	if options.BuildpackFile == "" {
		return errors.New("--buildpack-file is a required flag")
	}

	if options.PackageFile == "" {
		return errors.New("--package-file is a required flag")
	}

	bp, err := internal.ParseBuildpackConfig(options.BuildpackFile)
	if err != nil {
		return err
	}

	pkg, err := internal.ParsePackageConfig(options.PackageFile)
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

	err = internal.OverwriteBuildpackConfig(options.BuildpackFile, bp)
	if err != nil {
		return err
	}

	err = internal.OverwritePackageConfig(options.PackageFile, pkg)
	if err != nil {
		return err
	}

	return nil
}
