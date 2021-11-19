package packit

// BuildMetadata represents the build metadata details persisted in the
// build.toml file according to the buildpack lifecycle specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildtoml-toml.
type BuildMetadata struct {
	// BOM is the Bill-of-Material entries containing information about the
	// dependencies provided to the build environment.
	BOM []BOMEntry `toml:"bom"`

	SBOM SBOMEntries `toml:"-"`

	// Unmet is a list of unmet entries from the build process that it was unable
	// to provide.
	Unmet []UnmetEntry `toml:"unmet"`
}

func (b BuildMetadata) isEmpty() bool {
	return len(b.SBOM)+len(b.BOM)+len(b.Unmet) == 0
}
