package postal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type DependencyMapping struct {
	SHA256 string
	URI    string
}

func NewDependencyMapping(sha256, uri string) DependencyMapping {
	return DependencyMapping{
		SHA256: sha256,
		URI:    uri,
	}
}

// Given a target dependency, look up if there is a matching dependency mapping at the given binding path
func (d Dependency) FindDependencyMapping(bindingPath string) (DependencyMapping, error) {
	allBindings, err := filepath.Glob(filepath.Join(bindingPath, "*"))
	if err != nil {
		return NewDependencyMapping("", ""), err
	}

	for _, binding := range allBindings {
		// check if any of the bindings are of type "dependency-mapping"
		bindType, err := ioutil.ReadFile(filepath.Join(binding, "type"))
		if err != nil {
			return NewDependencyMapping("", ""), fmt.Errorf("couldn't read binding type: %w", err)
		}

		// if it is a dependency mapping, look for a SHA match
		if strings.Contains(string(bindType), "dependency-mapping") {
			if _, err := os.Stat(filepath.Join(binding, d.SHA256)); err != nil {
				if !os.IsNotExist(err) {
					return NewDependencyMapping("", ""), err
				}
				continue
			}

			// A matching SHA256 has been found, return the associated URI
			uri, err := ioutil.ReadFile(filepath.Join(binding, d.SHA256))
			if err != nil {
				return NewDependencyMapping("", ""), err
			}
			dependency := NewDependencyMapping(d.SHA256, strings.TrimSuffix(string(uri), "\n"))
			return dependency, nil
		}
	}
	return NewDependencyMapping("", ""), nil
}
