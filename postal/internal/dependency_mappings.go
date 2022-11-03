package internal

import (
	"fmt"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

//go:generate faux --interface BindingResolver --output fakes/binding_resolver.go
type BindingResolver interface {
	Resolve(typ, provider, platformDir string) ([]servicebindings.Binding, error)
}

type DependencyMappingResolver struct {
	bindingResolver BindingResolver
}

func NewDependencyMappingResolver(bindingResolver BindingResolver) DependencyMappingResolver {
	return DependencyMappingResolver{
		bindingResolver: bindingResolver,
	}
}

// FindDependencyMapping looks up if there is a matching dependency mapping
// If the binding is given in the form of `hash`, assume it is of algorithm `sha256`
// If the binding is given in the form of `algorithm:hash`, compare it to the full `checksum` input
func (d DependencyMappingResolver) FindDependencyMapping(checksum, platformDir string) (string, error) {
	bindings, err := d.bindingResolver.Resolve("dependency-mapping", "", platformDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve 'dependency-mapping' binding: %w", err)
	}

	hash := cargo.Checksum(checksum).Hash()

	for _, binding := range bindings {
		// binding provided in the form `hash` (no algorithm provided)
		// assumed to be of `sha256` algorithm
		if uri, ok := binding.Entries[hash]; ok && cargo.Checksum(checksum).Algorithm() == "sha256" {
			content, err := uri.ReadString()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(content), nil
			// binding provided in the form `algorithm:hash`
		} else if uri, ok := binding.Entries[checksum]; ok {
			content, err := uri.ReadString()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(content), nil
		}
	}

	return "", nil
}
