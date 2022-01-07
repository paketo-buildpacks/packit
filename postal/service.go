package postal

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/postal/internal"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/paketo-buildpacks/packit/v2/vacation"

	//nolint Ignore SA1019, usage of deprecated package within a deprecated test case
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
)

//go:generate faux --interface Transport --output fakes/transport.go

// Transport serves as the interface for types that can fetch dependencies
// given a location uri using either the http:// or file:// scheme.
type Transport interface {
	Drop(root, uri string) (io.ReadCloser, error)
}

//go:generate faux --interface MappingResolver --output fakes/mapping_resolver.go
// MappingResolver serves as the interface that looks up platform binding provided
// dependency mappings given a SHA256
type MappingResolver interface {
	FindDependencyMapping(SHA256, platformDir string) (string, error)
}

// Service provides a mechanism for resolving and installing dependencies given
// a Transport.
type Service struct {
	transport       Transport
	mappingResolver MappingResolver
}

// NewService creates an instance of a Service given a Transport.
func NewService(transport Transport) Service {
	return Service{
		transport: transport,
		mappingResolver: internal.NewDependencyMappingResolver(
			servicebindings.NewResolver(),
		),
	}
}

func (s Service) WithDependencyMappingResolver(mappingResolver MappingResolver) Service {
	s.mappingResolver = mappingResolver
	return s
}

// Resolve will pick the best matching dependency given a path to a
// buildpack.toml file, and the id, version, and stack value of a dependency.
// The version value is treated as a SemVer constraint and will pick the
// version that matches that constraint best. If the version is given as
// "default", the default version for the dependency with the given id will be
// used. If there is no default version for that dependency, a wildcard
// constraint will be used.
func (s Service) Resolve(path, id, version, stack string) (Dependency, error) {
	dependencies, defaultVersion, err := parseBuildpack(path, id)
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

	// Handle the pessmistic operator (~>)
	var re = regexp.MustCompile(`~>`)
	if re.MatchString(version) {
		res := re.ReplaceAllString(version, "")
		parts := strings.Split(res, ".")

		// if the version contains a major, minor, and patch use "~" Tilde Range Comparison
		// if the version contains a major and minor only, or a major version only use "^" Caret Range Comparison
		if len(parts) == 3 {
			version = "~" + res
		} else {
			version = "^" + res
		}
	}

	var compatibleVersions []Dependency
	versionConstraint, err := semver.NewConstraint(version)
	if err != nil {
		return Dependency{}, err
	}

	var supportedVersions []string
	for _, dependency := range dependencies {
		if dependency.ID != id || !stacksInclude(dependency.Stacks, stack) {
			continue
		}

		sVersion, err := semver.NewVersion(dependency.Version)
		if err != nil {
			return Dependency{}, err
		}

		if versionConstraint.Check(sVersion) {
			compatibleVersions = append(compatibleVersions, dependency)
		}

		supportedVersions = append(supportedVersions, dependency.Version)
	}

	if len(compatibleVersions) == 0 {
		return Dependency{}, fmt.Errorf(
			"failed to satisfy %q dependency version constraint %q: no compatible versions on %q stack. Supported versions are: [%s]",
			id,
			version,
			stack,
			strings.Join(supportedVersions, ", "),
		)
	}

	sort.Slice(compatibleVersions, func(i, j int) bool {
		iVersion := semver.MustParse(compatibleVersions[i].Version)
		jVersion := semver.MustParse(compatibleVersions[j].Version)
		return iVersion.GreaterThan(jVersion)
	})

	return compatibleVersions[0], nil
}

// Deliver will fetch and expand a dependency into a layer path location. The
// location of the CNBPath is given so that dependencies that may be included
// in a buildpack when packaged for offline consumption can be retrieved. If
// there is a dependency mapping for the specified dependency, Deliver will use
// the given dependency mapping URI to fetch the dependency. The dependency is
// validated against the checksum value provided on the Dependency and will
// error if there are inconsistencies in the fetched result.
func (s Service) Deliver(dependency Dependency, cnbPath, layerPath, platformPath string) error {
	dependencyMappingURI, err := s.mappingResolver.FindDependencyMapping(dependency.SHA256, platformPath)
	if err != nil {
		return fmt.Errorf("failure checking for dependency mappings: %s", err)
	}

	if dependencyMappingURI != "" {
		dependency.URI = dependencyMappingURI
	}

	bundle, err := s.transport.Drop(cnbPath, dependency.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch dependency: %s", err)
	}
	defer bundle.Close()

	validatedReader := cargo.NewValidatedReader(bundle, dependency.SHA256)

	name := dependency.Name
	if name == "" {
		name = filepath.Base(dependency.URI)
	}
	err = vacation.NewArchive(validatedReader).WithName(name).StripComponents(dependency.StripComponents).Decompress(layerPath)
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

// GenerateBillOfMaterials will generate a list of BOMEntry values given a
// collection of Dependency values.
//
// Deprecated: use sbom.GenerateFromDependency instead.
func (s Service) GenerateBillOfMaterials(dependencies ...Dependency) []packit.BOMEntry {
	var entries []packit.BOMEntry
	for _, dependency := range dependencies {
		paketoBomMetadata := paketosbom.BOMMetadata{
			Checksum: paketosbom.BOMChecksum{
				Algorithm: paketosbom.SHA256,
				Hash:      dependency.SHA256,
			},
			URI:     dependency.URI,
			Version: dependency.Version,
			Source: paketosbom.BOMSource{
				Checksum: paketosbom.BOMChecksum{
					Algorithm: paketosbom.SHA256,
					Hash:      dependency.SourceSHA256,
				},
				URI: dependency.Source,
			},
		}

		if dependency.CPE != "" {
			paketoBomMetadata.CPE = dependency.CPE
		}

		if (dependency.DeprecationDate != time.Time{}) {
			paketoBomMetadata.DeprecationDate = dependency.DeprecationDate
		}

		if dependency.Licenses != nil {
			paketoBomMetadata.Licenses = dependency.Licenses
		}

		if dependency.PURL != "" {
			paketoBomMetadata.PURL = dependency.PURL
		}

		entry := packit.BOMEntry{
			Name:     dependency.Name,
			Metadata: paketoBomMetadata,
		}

		entries = append(entries, entry)
	}

	return entries
}
