package sbom

import (
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

// Autobom will return a packit.SBOMFormatter containing the SBOM for the given path.
// It will also log information about the SBOM generation, including how long it took.
func Autobom(
	path string,
	SBOMFormats []string,
	clock chronos.Clock,
	logger scribe.Emitter) (packit.SBOMFormatter, error) {

	logger.GeneratingSBOM(path)
	var err error

	var sbomContent SBOM
	duration, err := clock.Measure(func() error {
		sbomContent, err = Generate(path)
		return err
	})
	if err != nil {
		return nil, err
	}
	logger.Action("Completed in %s", duration.Round(time.Millisecond))
	logger.Break()

	logger.FormattingSBOM(SBOMFormats...)

	return sbomContent.InFormats(SBOMFormats...)
}
