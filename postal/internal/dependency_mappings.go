package internal

import (
	"fmt"

	"github.com/paketo-buildpacks/packit/extensions"
)

type DependencyMappingResolver struct{}

func NewDependencyMappingResolver() DependencyMappingResolver {
	return DependencyMappingResolver{}
}

// Reference file structure for bindings directory
// - bindings
//    - some-binding
//       - type -> dependency-mapping
// 			 - some-sha -> some-uri
//       - other-sha -> other-uri

// Given a target dependency, look up if there is a matching dependency mapping at the given binding path
func (d DependencyMappingResolver) FindDependencyMapping(sha256, bindingPath string) (string, error) {
	bindings, err := extensions.ListServiceBindings(bindingPath)
	if err != nil {
		return "", fmt.Errorf("failed to list service bindings: %w", err)
	}

	for _, binding := range bindings {
		if binding.Type == "dependency-mapping" {
			uri, ok := binding.Secrets[sha256]
			if ok {
				return string(uri), nil
			}
		}
	}

	return "", nil
}
