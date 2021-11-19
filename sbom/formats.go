package sbom

type Format int

const (
	CycloneDXFormat Format = iota
	SPDXFormat
	SyftFormat
)
