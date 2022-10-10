package cargo

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ResolveDependency will pick the best matching dependency given a path to a
// buildpack.toml file, and the id, version, and stack value of a dependency.
// The version value is treated as a SemVer constraint and will pick the
// version that matches that constraint best. If the version is given as
// "default", the default version for the dependency with the given id will be
// used. If there is no default version for that dependency, a wildcard
// constraint will be used.
func ResolveDependency(config Config, id, version, stack string) (ConfigMetadataDependency, error) {
	if version == "" || version == "default" {
		version = "*"
		defaultVersion := config.Metadata.DefaultVersions[id]
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

	versionConstraint, err := semver.NewConstraint(version)
	if err != nil {
		return ConfigMetadataDependency{}, err
	}

	var compatibleVersions []ConfigMetadataDependency
	var supportedVersions []string
	for _, dependency := range config.Metadata.Dependencies {
		if dependency.ID != id {
			continue
		}

		if !(dependency.HasStack(stack) || dependency.HasStack("*")) {
			continue
		}

		sVersion, err := semver.NewVersion(dependency.Version)
		if err != nil {
			return ConfigMetadataDependency{}, err
		}

		if versionConstraint.Check(sVersion) {
			compatibleVersions = append(compatibleVersions, dependency)
		}

		supportedVersions = append(supportedVersions, dependency.Version)
	}

	if len(compatibleVersions) == 0 {
		return ConfigMetadataDependency{}, fmt.Errorf(
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
			return ConfigMetadataDependency{}, fmt.Errorf("multiple dependencies support wildcard stack for version: %q", version)
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

func stringSliceElementCount(slice []string, str string) int {
	count := 0
	for _, s := range slice {
		if s == str {
			count++
		}
	}

	return count
}

func stringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}
