package postal

import (
	"fmt"
	"io"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/vacation"
)

//go:generate faux --interface Transport --output fakes/transport.go
type Transport interface {
	Drop(root, uri string) (io.ReadCloser, error)
}

type Service struct {
	transport Transport
}

func NewService(transport Transport) Service {
	return Service{
		transport: transport,
	}
}

func (s Service) Resolve(path, name, version, stack string) (Dependency, error) {
	dependencies, defaultVersion, err := parseBuildpack(path, name)
	if err != nil {
		return Dependency{}, err
	}

	if version == "" {
		version = "default"
	}

	if version == "default" {
		version = "*"
		if defaultVersion != "" {
			version = defaultVersion
		}
	}

	var compatibleVersions []Dependency
	versionConstraint, err := semver.NewConstraint(version)
	if err != nil {
		return Dependency{}, err
	}

	for _, dependency := range dependencies {
		if dependency.ID != name || !dependency.Stacks.Include(stack) {
			continue
		}

		sVersion, err := semver.NewVersion(dependency.Version)
		if err != nil {
			return Dependency{}, err
		}

		if versionConstraint.Check(sVersion) {
			compatibleVersions = append(compatibleVersions, dependency)
		}
	}

	if len(compatibleVersions) == 0 {
		return Dependency{}, fmt.Errorf("failed to satisfy %q dependency version constraint %q: no compatible versions", name, version)
	}

	sort.Slice(compatibleVersions, func(i, j int) bool {
		iVersion := semver.MustParse(compatibleVersions[i].Version)
		jVersion := semver.MustParse(compatibleVersions[j].Version)
		return iVersion.GreaterThan(jVersion)
	})

	return compatibleVersions[0], nil
}

func (s Service) Install(dependency Dependency, cnbPath, layerPath string) error {
	bundle, err := s.transport.Drop(cnbPath, dependency.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch dependency: %s", err)
	}
	defer bundle.Close()

	validatedReader := cargo.NewValidatedReader(bundle, dependency.SHA256)

	err = vacation.NewTarGzipArchive(validatedReader).Decompress(layerPath)
	if err != nil {
		return err
	}

	ok, err := validatedReader.Valid()
	if err != nil {
		return fmt.Errorf("failed to validate dependency: %s", err)
	}

	if !ok {
		return fmt.Errorf("checksum does not match: %s", err)
	}

	return nil
}
