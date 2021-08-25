package packit

import "time"

// BOMEntry contains a bill of materials entry.
type BOMEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`

	// Metadata is the metadata of the entry.  Optional.
	// Metadata map[string]interface{} `toml:"metadata,omitempty"`
	Metadata *BOMMetadata `toml:"metadata,omitempty"`
}

type BOMMetadata struct {
	Architecture    string       `toml:"arch,omitempty"`
	CPE             string       `toml:"cpe,omitempty"`
	DeprecationDate *time.Time   `toml:"deprecation-date,omitempty"`
	Licenses        []string     `toml:"licenses,omitempty"`
	PURL            string       `toml:"purl,omitempty"`
	Checksum        *BOMChecksum `toml:"checksum,omitempty"`
	Summary         string       `toml:"summary,omitempty"`
	URI             string       `toml:"uri,omitempty"`
	Version         string       `toml:"version,omitempty"`
	Source          *BOMSource   `toml:"source,omitempty"`
}

type BOMChecksum struct {
	Algorithm string `toml:"algorithm,omitempty"`
	Hash      string `toml:"hash,omitempty"`
}

type BOMSource struct {
	Name            string       `toml:"name,omitempty"`
	Checksum        *BOMChecksum `toml:"checksum,omitempty"`
	UpstreamVersion string       `toml:"upstream-version,omitempty"`
	URI             string       `toml:"uri,omitempty"`
}

// UnmetEntry contains the name of an unmet dependency from the build process
type UnmetEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`
}
