package postal

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type Checksum = cargo.Checksum

// Dependency is a representation of a buildpack dependency.
type Dependency struct {
	// CPE is the Common Platform Enumerator for the dependency. Used in legacy
	// image label SBOM (GenerateBillOfMaterials).
	//
	// Deprecated: use CPEs instead.
	CPE string `toml:"cpe"`

	// CPEs are the Common Platform Enumerators for the dependency. Used in Syft
	// and SPDX JSON SBOMs. If unset, falls back to CPE.
	CPEs []string `toml:"cpes"`

	// DeprecationDate is the data upon which this dependency is considered deprecated.
	DeprecationDate time.Time `toml:"deprecation_date"`

	// Checksum is a string that includes an algorithm and the hex-encoded hash
	// of the built dependency separated by a colon. Example
	// sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.
	Checksum string `toml:"checksum"`

	// ID is the identifier used to specify the dependency.
	ID string `toml:"id"`

	// Licenses is a list of SPDX license identifiers of licenses in the dependency.
	Licenses []string `toml:"licenses"`

	// Name is the human-readable name of the dependency.
	Name string `toml:"name"`

	// PURL is the package URL for the dependency.
	PURL string `toml:"purl"`

	// SHA256 is the hex-encoded SHA256 checksum of the built dependency.
	//
	// Deprecated: use Checksum instead.
	SHA256 string `toml:"sha256"`

	// Source is the uri location of the source-code representation of the dependency.
	Source string `toml:"source"`

	// SourceChecksum is a string that includes an algorithm and the hex-encoded
	// hash of the source representation of the dependency separated by a colon.
	// Example sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.
	SourceChecksum string `toml:"source-checksum"`

	// SourceSHA256 is the hex-encoded SHA256 checksum of the source-code
	// representation of the dependency.
	//
	// Deprecated: use SourceChecksum instead.
	SourceSHA256 string `toml:"source_sha256"`

	// Stacks is a list of stacks for which the dependency is built.
	Stacks []string `toml:"stacks"`

	// URI is the uri location of the built dependency.
	URI string `toml:"uri"`

	// Version is the specific version of the dependency.
	Version string `toml:"version"`

	// StripComponents behaves like the --strip-components flag on tar command
	// removing the first n levels from the final decompression destination.
	StripComponents int `toml:"strip-components"`
}

func parseBuildpack(path, name string) ([]Dependency, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	var buildpack struct {
		Metadata struct {
			DefaultVersions map[string]string `toml:"default-versions"`
			Dependencies    []Dependency      `toml:"dependencies"`
		} `toml:"metadata"`
	}
	_, err = toml.NewDecoder(file).Decode(&buildpack)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	return buildpack.Metadata.Dependencies, buildpack.Metadata.DefaultVersions[name], nil
}

func stacksInclude(stacks []string, stack string) bool {
	for _, s := range stacks {
		if s == stack || s == "*" {
			return true
		}
	}
	return false
}

// Resolve will pick the best matching dependency given a path to a
// buildpack.toml file, and the id, version, and stack value of a dependency.
// The version value is treated as a SemVer constraint and will pick the
// version that matches that constraint best. If the version is given as
// "default", the default version for the dependency with the given id will be
// used. If there is no default version for that dependency, a wildcard
// constraint will be used.
func ResolveDependency(path, id, version, stack string) (Dependency, error) {
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
