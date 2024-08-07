// Package sbom implements standardized SBoM tooling that allows multiple SBoM
// formats to be generated from the same scanning information.
package sbom

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cpe"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/paketo-buildpacks/packit/v2/pexec"
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
				Packages:          catalog,
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
				Packages: catalog,
			},
			Source: source.Metadata{
				Scheme: source.DirectoryScheme,
				Path:   path,
			},
		},
	}, nil
}

func GenerateWithSyftCli(layersPath, layerName, scanDir string, mediaTypes ...string) error {

	args := []string{"scan", "-q"}
	for _, mediatype := range mediaTypes {
		sbomWriteLocation := filepath.Join(layersPath, fmt.Sprintf("%s.sbom.%s", layerName, getExtension(mediatype)))

		// TODO add @<version>
		args = append(args, "--output", fmt.Sprintf("%s=%s", sbomFormatToSyftOutputFormat(mediatype), sbomWriteLocation))
		// todo temporary
		fmt.Printf("Writing SBOM to %s\n", sbomWriteLocation)
	}

	args = append(args, fmt.Sprintf("dir:%s", scanDir))

	buffer := bytes.NewBuffer(nil)
	if err := pexec.NewExecutable("syft").Execute(pexec.Execution{
		Args:   args,
		Dir:    scanDir,
		Stdout: buffer,
		Stderr: buffer,
	}); err != nil {
		return fmt.Errorf("unable to run `syft %s`\n%w\n%s", args, err, buffer)
	}
	// todo remove
	fmt.Println("Finished syft command. output:")
	fmt.Println(buffer)
	fmt.Printf("args=%+v\n", args)

	// TODO clean cyclonedx file which has a timestamp and unique id which always change
	return nil
}

func getExtension(mediatype string) string {
	switch {
	case strings.HasPrefix(mediatype, CycloneDXFormat):
		return "cdx.json"
	case strings.HasPrefix(mediatype, SPDXFormat):
		return "spdx.json"
	// The syft tool does not support providing a version for its in-house standard.
	case mediatype == SyftFormat:
		return "syft.json"
	default:
		return ""
	}
}

func sbomFormatToSyftOutputFormat(mediatype string) string {
	fmt.Println("PPP: mediatype is " + mediatype)
	optionalVersionSegment := extractVersionSegment(mediatype)

	switch {
	case strings.HasPrefix(mediatype, CycloneDXFormat):
		return "cyclonedx-json" + optionalVersionSegment
	case strings.HasPrefix(mediatype, SPDXFormat):
		return "spdx-json" + optionalVersionSegment
	// The syft tool does not support providing a version for its in-house standard.
	case mediatype == SyftFormat:
		return "syft-json"
	default:
		return ""
	}
}

// look for a pattern like "application/vnd.cyclonedx+json;version=1.4"
func extractVersionSegment(input string) string {
	parts := strings.Split(input, ";")
	versionPart := parts[len(parts)-1]
	if strings.HasPrefix(versionPart, "version=") {
		return "@" + strings.Split(versionPart, "=")[1]
	}
	return ""
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
