// Package paketosbom implements a standardized SBoM format that can be used in
// Paketo Buildpacks.
//
// Deprecated: this package is frozen and will be removed in the next major
// release of packit.
package paketosbom

import (
	"fmt"
	"strings"
	"time"
)

// BOMMetadata represents how the Paketo-specific implementation of
// the Software Bill of Materials metadata components should be structured and named.
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

// GetBOMChecksumAlgorithm takes in an algorithm string, and reasonably tries
// to figure out the equivalent CycloneDX-supported algorithm field name.
// It returns an error if no reasonable supported format is found.
// Supported formats:
// { 'MD5'| 'SHA-1'| 'SHA-256'| 'SHA-384'| 'SHA-512'| 'SHA3-256'| 'SHA3-384'| 'SHA3-512'| 'BLAKE2b-256'| 'BLAKE2b-384'| 'BLAKE2b-512'| 'BLAKE3'}
func GetBOMChecksumAlgorithm(alg string) (algorithm, error) {
	for _, a := range []algorithm{SHA256, SHA1, SHA384, SHA512, SHA3256, SHA3384, SHA3512, BLAKE2B256, BLAKE2B384, BLAKE2B512, BLAKE3, MD5} {
		if strings.EqualFold(string(a), alg) || strings.EqualFold(strings.ReplaceAll(string(a), "-", ""), alg) {
			return a, nil
		}
	}

	return "", fmt.Errorf("failed to get supported BOM checksum algorithm: %s is not valid", alg)
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
