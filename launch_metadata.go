package packit

// LaunchMetadata represents the launch metadata details persisted in the
// launch.toml file according to the buildpack lifecycle specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#launchtoml-toml.
type LaunchMetadata struct {
	// Processes is a list of processes that will be returned to the lifecycle to
	// be executed during the launch phase.
	Processes []Process

	// Slices is a list of slices that will be returned to the lifecycle to be
	// exported as separate layers during the export phase.
	Slices []Slice

	// Labels is a map of key-value pairs that will be returned to the lifecycle to be
	// added as config label on the image metadata. Keys must be unique.
	Labels map[string]string

	// BOM is the Bill-of-Material entries containing information about the
	// dependencies provided to the launch environment.
	BOM []BOMEntry

	// SBOM is a type that implements SBOMFormatter and declares the formats that
	// bill-of-materials should be output for the launch SBoM.
	SBOM SBOMFormatter
}

func (l LaunchMetadata) isEmpty() bool {
	var sbom []SBOMFormat
	if l.SBOM != nil {
		sbom = l.SBOM.Formats()
	}

	return len(sbom)+len(l.Processes)+len(l.Slices)+len(l.Labels)+len(l.BOM) == 0
}
