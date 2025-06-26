package servicebindings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Binding represents metadata related to an external service.
type Binding struct {

	// Name is the name of the binding.
	Name string

	// Path is the path to the binding directory.
	Path string

	// Type is the type of the binding.
	Type string

	// Provider is the provider of the binding.
	Provider string

	// Entries is the set of entries that make up the binding.
	Entries map[string]*Entry
}

// Resolver resolves service bindings according to the kubernetes binding spec:
// https://github.com/k8s-service-bindings/spec#workload-projection.
//
// It also supports backwards compatibility with the legacy service binding spec:
// https://github.com/buildpacks/spec/blob/main/extensions/bindings.md
//
// It also supports reading bindings from VCAP_SERVICES
type Resolver struct{}

// NewResolver returns a new service binding resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve returns all bindings matching the given type and optional provider (case-insensitive). To match on type only,
// provider may be an empty string. Returns an error if there are problems loading bindings from the file system.
//
// The location of bindings is given by one of the following, in order of precedence:
//
//  1. SERVICE_BINDING_ROOT environment variable
//  2. CNB_BINDINGS environment variable, if above is not set
//  3. VCAP_SERVICES environment variable, if above is not set
//  4. `<platformDir>/bindings`, if all above are not set
func (r *Resolver) Resolve(typ, provider, platformDir string) ([]Binding, error) {
	bindings, err := loadBindings(platformDir)
	if err != nil {
		// return nil, fmt.Errorf("failed to load bindings: %w", err)
		return nil, err
	}

	var resolved []Binding
	for _, binding := range bindings {
		if (strings.EqualFold(binding.Type, typ)) &&
			(provider == "" || strings.EqualFold(binding.Provider, provider)) {
			resolved = append(resolved, binding)
		}
	}
	return resolved, nil
}

// ResolveOne returns a single binding matching the given type and optional provider (case-insensitive). To match on
// type only, provider may be an empty string. Returns an error if the number of matched bindings is not exactly one, or
// if there are problems loading bindings from the file system.
//
// The location of bindings is given by one of the following, in order of precedence:
//
//  1. SERVICE_BINDING_ROOT environment variable
//  2. CNB_BINDINGS environment variable, if above is not set
//  3. VCAP_SERVICES environment variable, if above is not set
//  4. `<platformDir>/bindings`, if all above are not set

func (r *Resolver) ResolveOne(typ, provider, platformDir string) (Binding, error) {
	bindings, err := r.Resolve(typ, provider, platformDir)
	if err != nil {
		return Binding{}, err
	}
	if len(bindings) != 1 {
		return Binding{}, fmt.Errorf("found %d bindings for type '%s' and provider '%s' but expected exactly 1", len(bindings), typ, provider)
	}
	return bindings[0], nil
}

func loadBindings(platformDir string) ([]Binding, error) {
	if path, ok := os.LookupEnv("SERVICE_BINDING_ROOT"); ok {
		return loadBindingsFromPath(path)
	}

	if path, ok := os.LookupEnv("CNB_BINDINGS"); ok {
		return loadBindingsFromPath(path)
	}

	if content, ok := os.LookupEnv("VCAP_SERVICES"); ok {
		return loadBindingsFromVcapServices(content)
	}

	return loadBindingsFromPath(filepath.Join(platformDir, "bindings"))
}

func loadBindingsFromPath(bindingRoot string) ([]Binding, error) {
	files, err := os.ReadDir(bindingRoot)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load bindings from '%s': %w", bindingRoot, err)
	}

	var bindings []Binding
	for _, file := range files {
		isLegacy, err := isLegacyBinding(bindingRoot, file.Name())
		if err != nil {
			return nil, err
		}

		var binding Binding
		if isLegacy {
			binding, err = loadLegacyBinding(bindingRoot, file.Name())
		} else {
			binding, err = loadK8sBinding(bindingRoot, file.Name())
		}
		if err != nil {
			return nil, fmt.Errorf("failed to load bindings from '%s': failed to read binding '%s': %w", bindingRoot, file.Name(), err)
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

// According to the legacy spec (https://github.com/buildpacks/spec/blob/main/extensions/bindings.md), a legacy binding
// has a `metadata` directory within the binding path.
func isLegacyBinding(bindingRoot, name string) (bool, error) {
	info, err := os.Stat(filepath.Join(bindingRoot, name, "metadata"))
	if err == nil {
		return info.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// See: https://github.com/k8s-service-bindings/spec#workload-projection
func loadK8sBinding(bindingRoot, name string) (Binding, error) {
	binding := Binding{
		Name:    name,
		Path:    filepath.Join(bindingRoot, name),
		Entries: map[string]*Entry{},
	}

	entries, err := loadEntries(filepath.Join(binding.Path))
	if err != nil {
		return Binding{}, err
	}

	typ, ok := entries["type"]
	if !ok {
		return Binding{}, errors.New("missing 'type'")
	}
	binding.Type, err = typ.ReadString()
	if err != nil {
		return Binding{}, err
	}
	binding.Type = strings.TrimSpace(binding.Type)
	delete(entries, "type")

	provider, ok := entries["provider"]
	if ok {
		binding.Provider, err = provider.ReadString()
		if err != nil {
			return Binding{}, err
		}
		binding.Provider = strings.TrimSpace(binding.Provider)
		delete(entries, "provider")
	}

	binding.Entries = entries

	return binding, nil
}

// See: https://github.com/buildpacks/spec/blob/main/extensions/bindings.md
func loadLegacyBinding(bindingRoot, name string) (Binding, error) {
	binding := Binding{
		Name:    name,
		Path:    filepath.Join(bindingRoot, name),
		Entries: map[string]*Entry{},
	}

	metadata, err := loadEntries(filepath.Join(binding.Path, "metadata"))
	if err != nil {
		return Binding{}, err
	}

	typ, ok := metadata["kind"]
	if !ok {
		return Binding{}, errors.New("missing 'kind'")
	}
	binding.Type, err = typ.ReadString()
	if err != nil {
		return Binding{}, err
	}
	binding.Type = strings.TrimSpace(binding.Type)
	delete(metadata, "kind")

	provider, ok := metadata["provider"]
	if !ok {
		return Binding{}, errors.New("missing 'provider'")
	}
	binding.Provider, err = provider.ReadString()
	if err != nil {
		return Binding{}, err
	}
	binding.Provider = strings.TrimSpace(binding.Provider)
	delete(metadata, "provider")

	binding.Entries = metadata

	secrets, err := loadEntries(filepath.Join(binding.Path, "secret"))
	if err != nil && !os.IsNotExist(err) {
		return Binding{}, err
	}
	if err == nil {
		for k, v := range secrets {
			binding.Entries[k] = v
		}
	}

	return binding, nil
}

func loadBindingsFromVcapServices(content string) ([]Binding, error) {
	var contentTyped map[string][]vcapServicesBinding

	err := json.Unmarshal([]byte(content), &contentTyped)
	if err != nil {
		return []Binding{}, fmt.Errorf("failed to load bindings from 'VCAP_SERVICES': %w", err)
	}

	bindings := []Binding{}
	for p, bArray := range contentTyped {
		for _, b := range bArray {
			entries := map[string]*Entry{}
			for k, v := range b.Credentials {
				entries[k], err = toJSONString(v)
				if err != nil {
					return nil, fmt.Errorf("failed to load bindings from 'VCAP_SERVICES': %w", err)
				}
			}

			bindingType := b.Label
			if value, ok := entries["type"]; ok {
				str, err := value.ReadString()
				if err != nil {
					return nil, fmt.Errorf("failed to load bindings from 'VCAP_SERVICES': could not read binding type: %w", err)
				}
				bindingType = str
			}

			bindings = append(bindings, Binding{
				Name:     b.Name,
				Type:     bindingType,
				Provider: p,
				Entries:  entries,
			})
		}
	}

	return bindings, nil
}

type vcapServicesBinding struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Credentials map[string]interface{} `json:"credentials"`
}

func toJSONString(input interface{}) (*Entry, error) {
	switch in := input.(type) {
	case string:
		return NewWithValue([]byte(in)), nil
	default:
		jsonProperty, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}
		return NewWithValue(jsonProperty), nil
	}
}

func loadEntries(path string) (map[string]*Entry, error) {
	entries := map[string]*Entry{}
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		entries[file.Name()] = NewEntry(filepath.Join(path, file.Name()))
	}
	return entries, nil
}
