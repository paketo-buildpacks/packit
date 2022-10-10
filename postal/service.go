package postal

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

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

// MappingResolver serves as the interface that looks up platform binding provided
// dependency mappings given a SHA256
//
//go:generate faux --interface MappingResolver --output fakes/mapping_resolver.go
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
//
// Deprecated: use cargo.ResolveDependency instead.
func (s Service) Resolve(path, id, version, stack string) (Dependency, error) {
	config, err := cargo.NewBuildpackParser().Parse(path)
	if err != nil {
		return Dependency{}, fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	dep, err := cargo.ResolveDependency(config, id, version, stack)
	if err != nil {
		return Dependency{}, err
	}

	return DependencyFrom(dep), nil
}

// Deliver will fetch and expand a dependency into a layer path location. The
// location of the CNBPath is given so that dependencies that may be included
// in a buildpack when packaged for offline consumption can be retrieved. If
// there is a dependency mapping for the specified dependency, Deliver will use
// the given dependency mapping URI to fetch the dependency. The dependency is
// validated against the checksum value provided on the Dependency and will
// error if there are inconsistencies in the fetched result.
func (s Service) DeliverDependency(dependency cargo.ConfigMetadataDependency, cnbPath, layerPath, platformPath string) error {
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

	var validatedReader cargo.ValidatedReader
	if dependency.SHA256 != "" {
		validatedReader = cargo.NewValidatedReader(bundle, fmt.Sprintf("sha256:%s", dependency.SHA256))
	} else {
		validatedReader = cargo.NewValidatedReader(bundle, dependency.Checksum)
	}

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
		return errors.New("failed to validate dependency: checksum does not match")
	}

	return nil
}

// Deliver will fetch and expand a dependency into a layer path location. The
// location of the CNBPath is given so that dependencies that may be included
// in a buildpack when packaged for offline consumption can be retrieved. If
// there is a dependency mapping for the specified dependency, Deliver will use
// the given dependency mapping URI to fetch the dependency. The dependency is
// validated against the checksum value provided on the Dependency and will
// error if there are inconsistencies in the fetched result.
//
// Deprecated: use DeliverDependency instead
func (s Service) Deliver(dependency Dependency, cnbPath, layerPath, platformPath string) error {
	return s.DeliverDependency(cargoDependencyFrom(dependency), cnbPath, layerPath, platformPath)
}

// GenerateBillOfMaterials will generate a list of BOMEntry values given a
// collection of Dependency values.
//
// Deprecated: use sbom.GenerateFromDependency instead.
func (s Service) GenerateBillOfMaterials(dependencies ...Dependency) []packit.BOMEntry {
	var entries []packit.BOMEntry
	for _, dependency := range dependencies {

		checksum := Checksum(dependency.SHA256)
		if len(dependency.Checksum) > 0 {
			checksum = Checksum(dependency.Checksum)
		}

		hash := checksum.Hash()
		paketoSbomAlgorithm, err := paketosbom.GetBOMChecksumAlgorithm(checksum.Algorithm())
		// GetBOMChecksumAlgorithm will set algorithm to UNKNOWN if there is an error
		if err != nil || hash == "" {
			paketoSbomAlgorithm = paketosbom.UNKNOWN
			hash = ""
		}

		sourceChecksum := Checksum(dependency.SourceSHA256)
		if len(dependency.Checksum) > 0 {
			sourceChecksum = Checksum(dependency.SourceChecksum)
		}

		sourceHash := sourceChecksum.Hash()
		paketoSbomSrcAlgorithm, err := paketosbom.GetBOMChecksumAlgorithm(sourceChecksum.Algorithm())
		// GetBOMChecksumAlgorithm will set algorithm to UNKNOWN if there is an error
		if err != nil || sourceHash == "" {
			paketoSbomSrcAlgorithm = paketosbom.UNKNOWN
			sourceHash = ""
		}

		paketoBomMetadata := paketosbom.BOMMetadata{
			Checksum: paketosbom.BOMChecksum{
				Algorithm: paketoSbomAlgorithm,
				Hash:      hash,
			},
			URI:     dependency.URI,
			Version: dependency.Version,
			Source: paketosbom.BOMSource{
				Checksum: paketosbom.BOMChecksum{
					Algorithm: paketoSbomSrcAlgorithm,
					Hash:      sourceHash,
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
