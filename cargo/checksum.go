package cargo

import (
	"strings"
)

// Checksum represents a checksum algorithm and hash pair formatted as
// algorithm:hash.
type Checksum string

// Algorithm returns the algorithm portion of the checksum string. If that
// portion is missing, it defaults to "sha256".
func (c Checksum) Algorithm() string {
	algorithm, _, found := strings.Cut(string(c), ":")
	if !found {
		return "sha256"
	}

	return algorithm
}

// Hash returns the hexidecimal encoded hash portion of the checksum string.
func (c Checksum) Hash() string {
	_, hash, found := strings.Cut(string(c), ":")
	if !found {
		hash = string(c)
	}

	return hash
}

// EqualTo returns true only when the given checksum algorithms and hashes
// match.
func (c Checksum) Match(o Checksum) bool {
	return strings.EqualFold(c.Algorithm(), o.Algorithm()) && c.Hash() == o.Hash()
}

// EqualTo returns true only when the given checksum formatted string
// algorithms and hashes match.
func (c Checksum) MatchString(o string) bool {
	return c.Match(Checksum(o))
}
