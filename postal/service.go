package postal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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

// MappingResolver serves as the interface that looks up platform binding provided
// dependency mappings given a SHA256
//
//go:generate faux --interface MappingResolver --output fakes/mapping_resolver.go
type MappingResolver interface {
	FindDependencyMapping(checksum, platformDir string) (string, error)
}

// MirrorResolver serves as the interface that looks for a dependency mirror via
// environment variable or binding
//
//go:generate faux --interface MirrorResolver --output fakes/mirror_resolver.go
type MirrorResolver interface {
	FindDependencyMirror(uri, platformDir string) (string, error)
}

// ErrNoDeps is a typed error indicating that no dependencies were resolved during Service.Resolve()
//
// errors can be tested against this type with: errors.As()
type ErrNoDeps struct {
	id                string
	version           string
	stack             string
	arch              string
	supportedVersions []string
}

// Error implements the error.Error interface
func (e *ErrNoDeps) Error() string {
	return fmt.Sprintf("failed to satisfy %q dependency version constraint %q: no compatible versions on %q stack with architecture %q. Supported versions are: [%s]",
		e.id,
		e.version,
		e.stack,
		e.arch,
		strings.Join(e.supportedVersions, ", "),
	)
}

// Service provides a mechanism for resolving and installing dependencies given
// a Transport.
type Service struct {
	transport       Transport
	mappingResolver MappingResolver
	mirrorResolver  MirrorResolver
}

// NewService creates an instance of a Service given a Transport.
func NewService(transport Transport) Service {
	return Service{
		transport: transport,
		mappingResolver: internal.NewDependencyMappingResolver(
			servicebindings.NewResolver(),
		),
		mirrorResolver: internal.NewDependencyMirrorResolver(
			servicebindings.NewResolver(),
		),
	}
}

func (s Service) WithDependencyMappingResolver(mappingResolver MappingResolver) Service {
	s.mappingResolver = mappingResolver
	return s
}

func (s Service) WithDependencyMirrorResolver(mirrorResolver MirrorResolver) Service {
	s.mirrorResolver = mirrorResolver
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

		if !isCorrectArch(dependency) {
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
		return Dependency{}, &ErrNoDeps{id, version, stack, archFromSystem(), supportedVersions}
	}

	stacksForVersion := map[string][]string{}

	for _, dep := range compatibleVersions {
		stacksForVersion[dep.Version] = append(stacksForVersion[dep.Version], dep.Stacks...)
	}

	for version, stacks := range stacksForVersion {
		count := stringSliceElementCount(stacks, "*")
		if count > 1 {
			return Dependency{}, fmt.Errorf("multiple dependencies support wildcard stack for version: %q", version)
		}
	}

	sort.Slice(compatibleVersions, func(i, j int) bool {
		iDep := compatibleVersions[i]
		jDep := compatibleVersions[j]

		jVersion := semver.MustParse(jDep.Version)
		iVersion := semver.MustParse(iDep.Version)

		if !iVersion.Equal(jVersion) {
			return iVersion.GreaterThan(jVersion)
		}

		iStacks := iDep.Stacks
		jStacks := jDep.Stacks

		// If either dependency supports the wildcard stack, it has lower
		// priority than a dependency that only supports a more specific stack.
		// This is true regardless of whether or not the dependency with
		// wildcard stack support also supports other stacks
		//
		// If is an error to have multiple dependencies with the same version
		// and wildcard stack support.
		// This is tested for above, and we would not enter this sort function
		// in this case

		if stringSliceContains(iStacks, "*") {
			return false
		}

		if stringSliceContains(jStacks, "*") {
			return true
		}

		// As mentioned above, this isn't a valid path to encounter because
		// only one dependency may have support for wildcard stacks for a given
		// version. We could panic, but it is preferable to return an invalid
		// sort order instead.
		//
		// This is untested as this path is not possible to encounter.
		return true
	})

	return compatibleVersions[0], nil
}

func isCorrectArch(dep Dependency) bool {
	systemArch := archFromSystem()

	if systemArch == "amd64" && dep.Arch == "" {
		return true
	}

	return systemArch == dep.Arch
}

func archFromSystem() string {
	archFromEnv, ok := os.LookupEnv("BP_ARCH")
	if !ok {
		archFromEnv = runtime.GOARCH
	}

	return archFromEnv
}

func stringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}

func stringSliceElementCount(slice []string, str string) int {
	count := 0
	for _, s := range slice {
		if s == str {
			count++
		}
	}

	return count
}

// Deliver will fetch and expand a dependency into a layer path location. The
// location of the CNBPath is given so that dependencies that may be included
// in a buildpack when packaged for offline consumption can be retrieved. If
// there is a dependency mapping for the specified dependency, Deliver will use
// the given dependency mapping URI to fetch the dependency. If there is a
// dependency mirror for the specified dependency, Deliver will use the mirror
// URI to fetch the dependency. If both a dependency mapping and mirror are BOTH
// present, the mapping will take precedence over the mirror.The dependency is
// validated against the checksum value provided on the Dependency and will error
// if there are inconsistencies in the fetched result.
func (s Service) Deliver(dependency Dependency, cnbPath, layerPath, platformPath string) error {
	dependencyChecksum := dependency.Checksum
	if dependency.SHA256 != "" {
		dependencyChecksum = fmt.Sprintf("sha256:%s", dependency.SHA256)
	}

	dependencyMirrorURI, err := s.mirrorResolver.FindDependencyMirror(dependency.URI, platformPath)
	if err != nil {
		return fmt.Errorf("failure checking for dependency mirror: %s", err)
	}

	dependencyMappingURI, err := s.mappingResolver.FindDependencyMapping(dependencyChecksum, platformPath)
	if err != nil {
		return fmt.Errorf("failure checking for dependency mappings: %s", err)
	}

	if dependencyMappingURI != "" {
		dependency.URI = dependencyMappingURI
	} else if dependencyMirrorURI != "" {
		dependency.URI = dependencyMirrorURI
	}

	bundle, err := s.transport.Drop(cnbPath, dependency.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch dependency: %s", err)
	}
	defer bundle.Close()

	validatedReader := cargo.NewValidatedReader(bundle, dependencyChecksum)

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
