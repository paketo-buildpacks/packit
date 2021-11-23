package sbom

type Format int

const (
	CycloneDXFormat Format = iota
	SPDXFormat
	SyftFormat
)

func (f Format) Extension() string {
	switch f {
	case CycloneDXFormat:
		return "cdx.json"
	case SPDXFormat:
		return "spdx.json"
	case SyftFormat:
		return "syft.json"
	}
	return ""
}
