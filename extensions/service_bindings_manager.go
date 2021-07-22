package extensions

import (
	"os"
	"path/filepath"
)

// ServiceBinding is a representation of a service binding that has
// been provided into the build container.
type ServiceBinding struct {
	// Name is the name of the directory that binding was contained within.
	Name string

	// Type is the content of the type file contained within the binding
	// directory.
	Type string

	// Provider is the content of the provider file contained within the binding
	// directory.
	Provider string

	// Path is the filepath of the binding directory.
	Path string

	// Secrets is a mapping of file name to contents for all other files
	// contained in the bindings directory.
	Secrets map[string][]byte
}

// ListServiceBindings returns all service bindings found in the given root
// directory. These bindings must conform the the Kubernetes Service Binding
// specification as documented in https://github.com/k8s-service-bindings/spec.
func ListServiceBindings(root string) ([]ServiceBinding, error) {
	directories, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		return nil, err
	}

	var bindings []ServiceBinding
	for _, directory := range directories {
		paths, err := filepath.Glob(filepath.Join(directory, "*"))
		if err != nil {
			return nil, err
		}

		binding := ServiceBinding{
			Name: filepath.Base(directory),
			Path: directory,
		}

		for _, path := range paths {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			base := filepath.Base(path)
			switch base {
			case "type":
				binding.Type = string(content)
			case "provider":
				binding.Provider = string(content)
			default:
				if binding.Secrets == nil {
					binding.Secrets = make(map[string][]byte)
				}

				binding.Secrets[base] = content
			}
		}

		bindings = append(bindings, binding)
	}

	return bindings, nil
}
