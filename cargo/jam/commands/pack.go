package commands

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

//go:generate faux --interface DirectoryDuplicator --output fakes/directory_duplicator.go
type DirectoryDuplicator interface {
	Duplicate(sourcePath, destPath string) error
}

//go:generate faux --interface DependencyCacher --output fakes/dependency_cacher.go
type DependencyCacher interface {
	Cache(root string, dependencies []cargo.ConfigMetadataDependency) ([]cargo.ConfigMetadataDependency, error)
}

type Pack struct {
	directoryDuplicator DirectoryDuplicator
	configParser        ConfigParser
	prePackager         PrePackager
	dependencyCacher    DependencyCacher
	tarBuilder          TarBuilder
	fileBundler         FileBundler
	stdout              io.Writer
}

func NewPack(
	directoryDuplicator DirectoryDuplicator,
	configParser ConfigParser,
	prePackager PrePackager,
	dependencyCacher DependencyCacher,
	fileBundler FileBundler,
	tarBuilder TarBuilder,
	stdout io.Writer,
) Pack {

	return Pack{
		directoryDuplicator: directoryDuplicator,
		configParser:        configParser,
		prePackager:         prePackager,
		dependencyCacher:    dependencyCacher,
		tarBuilder:          tarBuilder,
		fileBundler:         fileBundler,
		stdout:              stdout,
	}
}

func (p Pack) Execute(args []string) error {
	var (
		buildpackTOMLPath string
		output            string
		version           string
		offline           bool
		stack             string
	)

	fset := flag.NewFlagSet("pack", flag.ContinueOnError)
	fset.StringVar(&buildpackTOMLPath, "buildpack", "", "path to buildpack.toml")
	fset.StringVar(&output, "output", "", "path to location of output tarball")
	fset.StringVar(&version, "version", "", "version of the buildpack")
	fset.BoolVar(&offline, "offline", false, "enable offline caching of dependencies")
	fset.StringVar(&stack, "stack", "", "restricts dependencies to given stack")
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

	buildpackDir, err := ioutil.TempDir("", "dup-dest")
	if err != nil {
		return fmt.Errorf("unable to create temporary directory: %s", err)
	}
	defer os.RemoveAll(buildpackDir)

	err = p.directoryDuplicator.Duplicate(filepath.Dir(buildpackTOMLPath), buildpackDir)
	if err != nil {
		return fmt.Errorf("failed to duplicate directory: %s", err)
	}

	buildpackTOMLPath = filepath.Join(buildpackDir, filepath.Base(buildpackTOMLPath))

	config, err := p.configParser.Parse(buildpackTOMLPath)
	if err != nil {
		return fmt.Errorf("failed to parse buildpack.toml: %s", err)
	}

	config.Buildpack.Version = version

	fmt.Fprintf(p.stdout, "Packing %s %s...\n", config.Buildpack.Name, version)

	if stack != "" {
		var filteredDependencies []cargo.ConfigMetadataDependency
		for _, dep := range config.Metadata.Dependencies {
			if dep.HasStack(stack) {
				filteredDependencies = append(filteredDependencies, dep)
			}
		}

		config.Metadata.Dependencies = filteredDependencies
	}

	err = p.prePackager.Execute(config.Metadata.PrePackage, buildpackDir)
	if err != nil {
		return fmt.Errorf("failed to execute pre-packaging script %q: %s", config.Metadata.PrePackage, err)
	}

	if offline {
		config.Metadata.Dependencies, err = p.dependencyCacher.Cache(buildpackDir, config.Metadata.Dependencies)
		if err != nil {
			return fmt.Errorf("failed to cache dependencies: %s", err)
		}

		for _, dependency := range config.Metadata.Dependencies {
			config.Metadata.IncludeFiles = append(config.Metadata.IncludeFiles, strings.TrimPrefix(dependency.URI, "file:///"))
		}
	}

	files, err := p.fileBundler.Bundle(buildpackDir, config.Metadata.IncludeFiles, config)
	if err != nil {
		return fmt.Errorf("failed to bundle files: %s", err)
	}

	err = p.tarBuilder.Build(output, files)
	if err != nil {
		return fmt.Errorf("failed to create output: %s", err)
	}

	return nil
}
