package planning

// Metadata is the Paketo-buildpack specific data included in build plan
// requirements.
// See https://github.com/buildpacks/spec/blob/main/buildpack.md#build-plan-toml
type Metadata struct {
	// Version requests a specific version of the dependency being required
	Version string `toml:"version"`
	// VersionSource indicates the reason why this specific Version is being required.
	// It can be used by the buildpack providing the dependency to determine which
	// version request has the highest priority.
	VersionSource string `toml:"version-source"`
	// Build indicates that the providing buildpack should make the dependency available
	// in a build layer
	Build bool `toml:"build"`
	// Launch indicates that the providing buildpack should make the dependency available
	// in a launch layer
	Launch bool `toml:"launch"`
}

func NewMetadata(input map[string]interface{}) (metadata Metadata) {
	if version, ok := input["version"].(string); ok {
		metadata.Version = version
	}

	if versionSource, ok := input["version-source"].(string); ok {
		metadata.VersionSource = versionSource
	}

	if build, ok := input["build"].(bool); ok {
		metadata.Build = build
	}

	if launch, ok := input["launch"].(bool); ok {
		metadata.Launch = launch
	}

	return
}

func (metadata Metadata) ToMap() (result map[string]interface{}) {
	result = make(map[string]interface{})
	result["version"] = metadata.Version
	result["version-source"] = metadata.VersionSource
	result["build"] = metadata.Build
	result["launch"] = metadata.Launch

	return
}
