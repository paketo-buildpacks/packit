package packit

import "time"

// BOMEntry contains a bill of materials entry.
type BOMEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`

	// Metadata is the metadata of the entry.  Optional.
	// Metadata map[string]interface{} `toml:"metadata,omitempty"`
	Metadata *BomMetadata `toml:"metadata,omitempty"`
}

type BomMetadata struct {
	Architecture    string       `toml:"arch,omitempty"`
	CPE             string       `toml:"cpe,omitempty"`
	DeprecationDate *time.Time   `toml:"deprecation-date,omitempty"`
	Licenses        []string     `toml:"licenses,omitempty"`
	PURL            string       `toml:"purl,omitempty"`
	Checksum        *BomChecksum `toml:"checksum,omitempty"`
	Summary         string       `toml:"summary,omitempty"`
	URI             string       `toml:"uri,omitempty"`
	Version         string       `toml:"version,omitempty"`
	Source          *BomSource   `toml:"source,omitempty"`
}

type BomChecksum struct {
	Algorithm string `toml:"alg,omitempty"`
	Hash      string `toml:"hash,omitempty"`
}

type BomSource struct {
	Name            string       `toml:"name,omitempty"`
	Checksum        *BomChecksum `toml:"checksum,omitempty"`
	UpstreamVersion string       `toml:"upstream-version,omitempty"`
	URI             string       `toml:"uri,omitempty"`
}

// UnmetEntry contains the name of an unmet dependency from the build process
type UnmetEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`
}
