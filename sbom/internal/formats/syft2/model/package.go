package model

import (
	"encoding/json"
	"fmt"

	"github.com/anchore/syft/syft/file"
	"github.com/anchore/syft/syft/license"
	"github.com/anchore/syft/syft/pkg"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/log"
)

// Package represents a pkg.Package object specialized for JSON marshaling and unmarshalling.
type Package struct {
	PackageBasicData
	PackageCustomData
}

// PackageBasicData contains non-ambiguous values (type-wise) from pkg.Package.
type PackageBasicData struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Version   string          `json:"version"`
	Type      pkg.Type        `json:"type"`
	FoundBy   string          `json:"foundBy"`
	Locations []file.Location `json:"locations"`
	Licenses  licenses        `json:"licenses"`
	Language  pkg.Language    `json:"language"`
	CPEs      []string        `json:"cpes"`
	PURL      string          `json:"purl"`
}

type licenses []License

type License struct {
	Value          string          `json:"value"`
	SPDXExpression string          `json:"spdxExpression"`
	Type           license.Type    `json:"type"`
	URLs           []string        `json:"urls"`
	Locations      []file.Location `json:"locations"`
}

func newModelLicensesFromValues(licenses []string) (ml []License) {
	for _, v := range licenses {
		expression, err := license.ParseExpression(v)
		if err != nil {
			log.Trace("could not find valid spdx expression for %s: %w", v, err)
		}
		ml = append(ml, License{
			Value:          v,
			SPDXExpression: expression,
			Type:           license.Declared,
		})
	}
	return ml
}

func (f *licenses) UnmarshalJSON(b []byte) error {
	var licenses []License
	if err := json.Unmarshal(b, &licenses); err != nil {
		var simpleLicense []string
		if err := json.Unmarshal(b, &simpleLicense); err != nil {
			return fmt.Errorf("unable to unmarshal license: %w", err)
		}
		licenses = newModelLicensesFromValues(simpleLicense)
	}
	*f = licenses
	return nil
}

// PackageCustomData contains ambiguous values (type-wise) from pkg.Package.
type PackageCustomData struct {
	MetadataType pkg.MetadataType `json:"metadataType"`
	Metadata     interface{}      `json:"metadata"`
}

// packageMetadataUnpacker is all values needed from Package to disambiguate ambiguous fields during json unmarshaling.
type packageMetadataUnpacker struct {
	MetadataType pkg.MetadataType `json:"metadataType"`
	Metadata     json.RawMessage  `json:"metadata"`
}

func (p *packageMetadataUnpacker) String() string {
	return fmt.Sprintf("metadataType: %s, metadata: %s", p.MetadataType, string(p.Metadata))
}

// UnmarshalJSON is a custom unmarshaller for handling basic values and values with ambiguous types.
// nolint:funlen
func (p *Package) UnmarshalJSON(b []byte) error {
	var basic PackageBasicData
	if err := json.Unmarshal(b, &basic); err != nil {
		return err
	}
	p.PackageBasicData = basic

	var unpacker packageMetadataUnpacker
	if err := json.Unmarshal(b, &unpacker); err != nil {
		// log.Warnf("failed to unmarshall into packageMetadataUnpacker: %v", err)
		return err
	}

	p.MetadataType = unpacker.MetadataType

	switch p.MetadataType {
	case pkg.ApkMetadataType:
		var payload pkg.ApkMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.RpmMetadataType:
		var payload pkg.RpmMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.DpkgMetadataType:
		var payload pkg.DpkgMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.JavaMetadataType:
		var payload pkg.JavaMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.RustCargoPackageMetadataType:
		var payload pkg.CargoPackageMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.GemMetadataType:
		var payload pkg.GemMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.KbPackageMetadataType:
		var payload pkg.KbPackageMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.PythonPackageMetadataType:
		var payload pkg.PythonPackageMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	case pkg.NpmPackageJSONMetadataType:
		var payload pkg.NpmPackageJSONMetadata
		if err := json.Unmarshal(unpacker.Metadata, &payload); err != nil {
			return err
		}
		p.Metadata = payload
	}

	return nil
}
