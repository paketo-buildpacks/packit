package postal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type DependencyMappingResolver struct{}

func NewDependencyMappingResolver() DependencyMappingResolver {
	return DependencyMappingResolver{}
}

// Given a target dependency, look up if there is a matching dependency mapping at the given binding path
func (d DependencyMappingResolver) FindDependencyMapping(sha256, bindingPath string) (string, error) {
	fmt.Println(bindingPath)
	allBindings, err := filepath.Glob(filepath.Join(bindingPath, "*"))
	if err != nil {
		return "", err
	}

	for _, binding := range allBindings {
		// check if any of the bindings are of type "dependency-mapping"
		bindType, err := ioutil.ReadFile(filepath.Join(binding, "type"))
		if err != nil {
			return "", fmt.Errorf("couldn't read binding type: %w", err)
		}

		// if it is a dependency mapping, look for a SHA match
		if strings.Contains(string(bindType), "dependency-mapping") {
			if _, err := os.Stat(filepath.Join(binding, sha256)); err != nil {
				if !os.IsNotExist(err) {
					return "", err
				}
				continue
			}

			// A matching SHA256 has been found, return the associated URI
			uri, err := ioutil.ReadFile(filepath.Join(binding, sha256))
			if err != nil {
				return "", err
			}
			fmt.Println("got the good uri")
			return strings.TrimSuffix(string(uri), "\n"), nil
		}
	}
	return "", nil
}
