package packit

// BuildpackInfo is a representation of the basic information for a buildpack
// provided in its buildpack.toml file as described in the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildpacktoml-toml.
type BuildpackInfo struct {
	// ID is the identifier specified in the `buildpack.id` field of the
	// buildpack.toml.
	ID string `toml:"id"                    json:"id,omitempty"`

	// Name is the identifier specified in the `buildpack.name` field of the
	// buildpack.toml.
	Name string `toml:"name"                  json:"name,omitempty"`

	// Version is the identifier specified in the `buildpack.version` field of
	// the buildpack.toml.
	Version string `toml:"version"               json:"version,omitempty"`

	// Homepage is the identifier specified in the `buildpack.homepage` field of
	// the buildpack.toml.
	Homepage string `toml:"homepage,omitempty"    json:"homepage,omitempty"`

	// ClearEnv is the identifier specificed in the `buildpack.clen-env` field of
	// the buildpack.toml.
	ClearEnv bool `toml:"clear-env,omitempty"   json:"clear-env,omitempty"`

	// Description is the identifier specified in the `buildpack.description`
	// field of the buildpack.toml.
	Description string `toml:"description,omitempty" json:"description,omitempty"`

	// Keywords are the identifiers specified in the `buildpack.keywords` field
	// of the buildpack.toml.
	Keywords []string `toml:"keywords,omitempty"    json:"keywords,omitempty"`

	// Licenses are the list of licenses specified in the `buildpack.licenses`
	// fields of the buildpack.toml.
	Licenses []BuildpackInfoLicense `toml:"licenses,omitempty"    json:"licenses,omitempty"`

	// SBOMFormats is the list of Software Bill of Materials media types that the buildpack
	// produces (e.g. "application/spdx+json").
	SBOMFormats []string `toml:"sbom-formats,omitempty"    json:"sbom-formats,omitempty"`
}

// BuildpackInfoLicense is a representation of a license specified in the
// buildpack.toml as described in the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildpacktoml-toml.
type BuildpackInfoLicense struct {
	// Type is the identifier specified in the `buildpack.licenses.type` field of
	// the buildpack.toml.
	Type string `toml:"type" json:"type"`

	// URI is the identifier specified in the `buildpack.licenses.uri` field of
	// the buildpack.toml.
	URI string `toml:"uri"  json:"uri"`
}
