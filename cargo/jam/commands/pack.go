package commands

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/cloudfoundry/packit/cargo"
)

//go:generate faux --interface ConfigParser --output fakes/config_parser.go
type ConfigParser interface {
	Parse(path string) (cargo.Config, error)
}

//go:generate faux --interface FileBundler --output fakes/file_bundler.go
type FileBundler interface {
	Bundle(path string, files []string, config cargo.Config) ([]cargo.File, error)
}

//go:generate faux --interface TarBuilder --output fakes/tar_builder.go
type TarBuilder interface {
	Build(path string, files []cargo.File) error
}

//go:generate faux --interface PrePackager --output fakes/prepackager.go
type PrePackager interface {
	Execute(path, rootDir string) error
}

type Pack struct {
	configParser ConfigParser
	tarBuilder   TarBuilder
	fileBundler  FileBundler
	prePackager  PrePackager
	stdout       io.Writer
}

func NewPack(configParser ConfigParser, prePackager PrePackager, fileBundler FileBundler, tarBuilder TarBuilder, stdout io.Writer) Pack {
	return Pack{
		configParser: configParser,
		tarBuilder:   tarBuilder,
		fileBundler:  fileBundler,
		prePackager:  prePackager,
		stdout:       stdout,
	}
}

func (p Pack) Execute(args []string) error {
	var (
		buildpackTOMLPath string
		output            string
		version           string
	)

	fset := flag.NewFlagSet("pack", flag.ContinueOnError)
	fset.StringVar(&buildpackTOMLPath, "buildpack", "", "path to buildpack.toml")
	fset.StringVar(&output, "output", "", "path to location of output tarball")
	fset.StringVar(&version, "version", "", "version of the buildpack")
	err := fset.Parse(args)
	if err != nil {
		return err
	}

	if buildpackTOMLPath == "" {
		return errors.New("missing required flag --buildpack")
	}

	if output == "" {
		return errors.New("missing required flag --output")
	}

	if version == "" {
		return errors.New("missing required flag --version")
	}

	config, err := p.configParser.Parse(buildpackTOMLPath)
	if err != nil {
		return fmt.Errorf("failed to parse buildpack.toml: %s", err)
	}

	config.Buildpack.Version = version

	err = p.prePackager.Execute(config.Metadata.PrePackage, filepath.Dir(buildpackTOMLPath))
	if err != nil {
		return fmt.Errorf("failed to execute pre-packaging script %q: %s", config.Metadata.PrePackage, err)
	}

	fmt.Fprintf(p.stdout, "Packing %s %s...\n", config.Buildpack.Name, version)

	files, err := p.fileBundler.Bundle(filepath.Dir(buildpackTOMLPath), config.Metadata.IncludeFiles, config)
	if err != nil {
		return fmt.Errorf("failed to bundle files: %s", err)
	}

	err = p.tarBuilder.Build(output, files)
	if err != nil {
		return fmt.Errorf("failed to create output: %s", err)
	}

	return nil
}
