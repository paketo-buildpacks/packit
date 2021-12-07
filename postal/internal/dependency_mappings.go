package internal

import (
	"fmt"

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
func (d DependencyMappingResolver) FindDependencyMapping(sha256, platformDir string) (string, error) {
	bindings, err := d.bindingResolver.Resolve("dependency-mapping", "", platformDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve 'dependency-mapping' binding: %w", err)
	}

	for _, binding := range bindings {
		if uri, ok := binding.Entries[sha256]; ok {
			return uri.ReadString()
		}
	}

	return "", nil
}
