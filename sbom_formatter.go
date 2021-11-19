package packit

import "io"

// SBOMFormat represents the mapping of a formatted SBOM content to its file
// extension on the filesystem.
type SBOMFormat struct {
	Extension string
	Content   io.Reader
}

// SBOMFormatter defines the interface for types capable of generating
// formatted SBoMs.
type SBOMFormatter interface {
	Formats() []SBOMFormat
}

// SBOMFormats implements the SBOMFormatter interface by wrapping a slice of
// SBOMFormat instances. This allows for quick, inline instantiation of a type
// that conforms to the SBOMFormatter interface.
type SBOMFormats []SBOMFormat

// Formats returns the slice of SBOMFormat instances wrapped by the SBOMFormats
// type.
func (f SBOMFormats) Formats() []SBOMFormat {
	return f
}
