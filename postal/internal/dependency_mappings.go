package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	allBindings, err := filepath.Glob(filepath.Join(bindingPath, "*"))
	if err != nil {
		return "", err
	}

	for _, binding := range allBindings {
		bindType, err := os.ReadFile(filepath.Join(binding, "type"))
		if err != nil {
			return "", fmt.Errorf("couldn't read binding type: %w", err)
		}

		if strings.TrimSpace(string(bindType)) == "dependency-mapping" {
			if _, err := os.Stat(filepath.Join(binding, sha256)); err != nil {
				if !os.IsNotExist(err) {
					return "", err
				}
				continue
			}

			uri, err := os.ReadFile(filepath.Join(binding, sha256))
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(uri)), nil
		}
	}
	return "", nil
}
