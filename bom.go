package packit

import (
	"time"
)

// BOMEntry contains a bill of materials entry.
type BOMEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`

	// Metadata is the metadata of the entry.  Optional.
	Metadata BOMMetadata `toml:"metadata,omitempty"`
}

type BOMMetadata struct {
	Architecture    string      `toml:"arch,omitempty"`
	CPE             string      `toml:"cpe,omitempty"`
	DeprecationDate time.Time   `toml:"deprecation-date,omitempty"`
	Licenses        []string    `toml:"licenses,omitempty"`
	PURL            string      `toml:"purl,omitempty"`
	Checksum        BOMChecksum `toml:"checksum,omitempty"`
	Summary         string      `toml:"summary,omitempty"`
	URI             string      `toml:"uri,omitempty"`
	Version         string      `toml:"version,omitempty"`
	Source          BOMSource   `toml:"source,omitempty"`
}

type BOMSource struct {
	Name            string      `toml:"name,omitempty"`
	Checksum        BOMChecksum `toml:"checksum,omitempty"`
	UpstreamVersion string      `toml:"upstream-version,omitempty"`
	URI             string      `toml:"uri,omitempty"`
}

type BOMChecksum struct {
	Algorithm ChecksumAlgorithm `toml:"algorithm,omitempty"`
	Hash      string            `toml:"hash,omitempty"`
}

type ChecksumAlgorithm interface {
	alg() algorithm
}

type algorithm string

func (a algorithm) alg() algorithm {
	return a
}

const (
	SHA256     algorithm = "SHA-256"
	SHA1       algorithm = "SHA-1"
	SHA384     algorithm = "SHA-384"
	SHA512     algorithm = "SHA-512"
	SHA3256    algorithm = "SHA3-256"
	SHA3384    algorithm = "SHA3-384"
	SHA3512    algorithm = "SHA3-512"
	BLAKE2B256 algorithm = "BLAKE2b-256"
	BLAKE2B384 algorithm = "BLAKE2b-384"
	BLAKE2B512 algorithm = "BLAKE2b-512"
	BLAKE3     algorithm = "BLAKE3"
	MD5        algorithm = "MD5"
)

// UnmetEntry contains the name of an unmet dependency from the build process
type UnmetEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`
}
