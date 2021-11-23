package sbom

import (
	"io"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/paketo-buildpacks/packit/postal"
)

type SBOM struct {
	syft sbom.SBOM
}

func Generate(path string) (SBOM, error) {
	src, err := source.NewFromDirectory(path)
	if err != nil {
		panic(err)
	}

	catalog, _, distro, err := syft.CatalogPackages(&src, source.UnknownScope)
	if err != nil {
		panic(err)
	}

	return SBOM{
		syft: sbom.SBOM{
			Artifacts: sbom.Artifacts{
				PackageCatalog: catalog,
				Distro:         distro,
			},
			Source: src.Metadata,
		},
	}, nil
}

func GenerateFromDependency(dependency postal.Dependency, path string) (SBOM, error) {
	cpe, err := pkg.NewCPE(dependency.CPE)
	if err != nil {
		panic(err)
	}

	catalog := pkg.NewCatalog(pkg.Package{
		Name:     dependency.Name,
		Version:  dependency.Version,
		Licenses: dependency.Licenses,
		CPEs:     []pkg.CPE{cpe},
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

func (s SBOM) Format(format Format) io.Reader {
	switch format {
	case CycloneDXFormat:
		return NewCycloneDXFormatter(s)
	case SPDXFormat:
		return NewSPDXFormatter(s)
	case SyftFormat:
		return NewSyftFormatter(s)
	default:
		return nil
	}
}

func (s SBOM) IsEmpty() bool {
	return s.syft.Artifacts.PackageCatalog.PackageCount() == 0
}
