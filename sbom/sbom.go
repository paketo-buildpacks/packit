// Package sbom implements standardized SBoM tooling that allows multiple SBoM
// formats to be generated from the same scanning information.
package sbom

import (
	"fmt"
	"os"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cpe"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

// UnknownCPE is a Common Platform Enumeration (CPE) that uses the NA (Not
// applicable) logical operator for all components of its name. It is designed
// not to match with other CPEs, to avoid false positive CPE matches.
const UnknownCPE = "cpe:2.3:-:-:-:-:-:-:-:-:-:-:-"

// SBOM holds the internal representation of the generated software
// bill-of-materials. This type can be combined with a FormattedReader to
// output the SBoM in a number of file formats.
type SBOM struct {
	syft sbom.SBOM
}

func NewSBOM(syft sbom.SBOM) SBOM {
	return SBOM{syft: syft}
}

// Generate returns a populated SBOM given a path to a directory to scan.
func Generate(path string) (SBOM, error) {
	info, err := os.Stat(path)
	if err != nil {
		return SBOM{}, err
	}

	var src source.Source
	if info.IsDir() {
		src, err = source.NewFromDirectory(path)
		if err != nil {
			return SBOM{}, err
		}
	} else {
		var cleanup func()
		src, cleanup = source.NewFromFile(path)
		defer cleanup()
	}

	config := cataloger.Config{
		Search: cataloger.SearchConfig{
			Scope: source.UnknownScope,
		},
	}

	catalog, _, release, err := syft.CatalogPackages(&src, config)
	if err != nil {
		return SBOM{}, err
	}

	return SBOM{
		syft: sbom.SBOM{
			Artifacts: sbom.Artifacts{
				PackageCatalog:    catalog,
				LinuxDistribution: release,
			},
			Source: src.Metadata,
		},
	}, nil
}

// GenerateFromDependency returns a populated SBOM given a postal.Dependency
// and the directory path where the dependency will be located within the
// application image.

//nolint Ignore SA1019, informed usage of deprecated package
func GenerateFromDependency(dependency postal.Dependency, path string) (SBOM, error) {

	//nolint Ignore SA1019, informed usage of deprecated package
	if dependency.CPE == "" {
		dependency.CPE = UnknownCPE
	}
	if len(dependency.CPEs) == 0 {
		//nolint Ignore SA1019, informed usage of deprecated package
		dependency.CPEs = []string{dependency.CPE}
	}

	var cpes []cpe.CPE
	for _, cpeString := range dependency.CPEs {
		cpe, err := cpe.New(cpeString)
		if err != nil {
			return SBOM{}, err
		}
		cpes = append(cpes, cpe)
	}

	catalog := pkg.NewCatalog(pkg.Package{
		Name:     dependency.Name,
		Version:  dependency.Version,
		Licenses: dependency.Licenses,
		CPEs:     cpes,
		PURL:     dependency.PURL,
	})

	return SBOM{
		syft: sbom.SBOM{
			Artifacts: sbom.Artifacts{
				PackageCatalog: catalog,
			},
			Source: source.Metadata{
				Scheme: source.DirectoryScheme,
				Path:   path,
			},
		},
	}, nil
}

// InFormats returns a Formatter containing mappings for the given Formats.
func (s SBOM) InFormats(mediaTypes ...string) (Formatter, error) {
	var fs []sbom.FormatID
	for _, m := range mediaTypes {
		format, err := sbomFormatByMediaType(m)
		if err != nil {
			return Formatter{}, err
		}

		if format.Extension() == "" {
			return Formatter{}, fmt.Errorf("unable to determine file extension for SBOM format '%s'", format.ID())
		}

		fs = append(fs, format.ID())
	}

	return Formatter{sbom: s, formatIDs: fs}, nil
}
