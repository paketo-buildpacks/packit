package internal

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type DependencyMirrorResolver struct {
	bindingResolver BindingResolver
}

func NewDependencyMirrorResolver(bindingResolver BindingResolver) DependencyMirrorResolver {
	return DependencyMirrorResolver{
		bindingResolver: bindingResolver,
	}
}

func formatAndVerifyMirror(mirror, uri string) (string, error) {
	mirrorURL, err := url.Parse(mirror)
	if err != nil {
		return "", err
	}

	uriURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if mirrorURL.Scheme != "https" && mirrorURL.Scheme != "file" {
		return "", fmt.Errorf("invalid mirror scheme")
	}

	mirrorURL.Path = strings.Replace(mirrorURL.Path, "{originalHost}", uriURL.Host+uriURL.Path, 1)
	return mirrorURL.String(), nil
}

func (d DependencyMirrorResolver) FindDependencyMirror(uri, platformDir string) (string, error) {
	mirror, err := d.findMirrorFromEnv(uri)
	if err != nil {
		return "", err
	}

	if mirror != "" {
		return formatAndVerifyMirror(mirror, uri)
	}

	mirror, err = d.findMirrorFromBinding(uri, platformDir)
	if err != nil {
		return "", err
	}

	if mirror != "" {
		return formatAndVerifyMirror(mirror, uri)
	}

	return "", nil
}

func (d DependencyMirrorResolver) findMirrorFromEnv(uri string) (string, error) {
	const DefaultMirror = "BP_DEPENDENCY_MIRROR"
	const NonDefaultMirrorPrefix = "BP_DEPENDENCY_MIRROR_"
	mirrors := make(map[string]string)
	environmentVariables := os.Environ()
	for _, ev := range environmentVariables {
		pair := strings.SplitN(ev, "=", 2)
		key := pair[0]
		value := pair[1]

		if !strings.Contains(key, DefaultMirror) {
			continue
		}

		if key == DefaultMirror {
			mirrors["default"] = value
			continue
		}

		// convert key
		hostname := strings.SplitN(key, NonDefaultMirrorPrefix, 2)[1]
		hostname = strings.ReplaceAll(strings.ReplaceAll(hostname, "__", "-"), "_", ".")
		hostname = strings.ToLower(hostname)
		mirrors[hostname] = value

		if !strings.Contains(uri, hostname) {
			continue
		}

		return value, nil
	}

	if mirrorUri, ok := mirrors["default"]; ok {
		return mirrorUri, nil
	}

	return "", nil
}

func (d DependencyMirrorResolver) findMirrorFromBinding(uri, platformDir string) (string, error) {
	bindings, err := d.bindingResolver.Resolve("dependency-mirror", "", platformDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve 'dependency-mirror' binding: %w", err)
	}

	if len(bindings) > 1 {
		return "", fmt.Errorf("cannot have multiple bindings of type 'dependency-mirror'")
	}

	if len(bindings) == 0 {
		return "", nil
	}

	mirror := ""
	entries := bindings[0].Entries
	for hostname, entry := range entries {
		if hostname == "default" {
			mirror, err = entry.ReadString()
			if err != nil {
				return "", err
			}
			continue
		}

		if !strings.Contains(uri, hostname) {
			continue
		}

		mirror, err = entry.ReadString()
		if err != nil {
			return "", err
		}

		return mirror, nil
	}

	return mirror, nil
}
