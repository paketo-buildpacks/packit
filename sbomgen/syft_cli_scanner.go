package sbomgen

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) (err error)
}

// SyftCLIScanner implements scanning a path using the `syft` CLI
// to generate SBOM, process it, and write it to a location that complies with
// the buildpacks spec. Supports CycloneDX, SPDX and Syft mediatypes, with an
// optional version param for CycloneDX and Syft.
//
// Example Usage:
//
// syftCLIScanner := sbomgen.NewSyftCLIScanner(
//	pexec.NewExecutable("syft"),
//	scribe.NewEmitter(os.Stdout),
// )
type SyftCLIScanner struct {
	syftCLI Executable
	logger  scribe.Emitter
}

func NewSyftCLIScanner(syftCLI Executable, logger scribe.Emitter) SyftCLIScanner {
	return SyftCLIScanner{
		syftCLI: syftCLI,
		logger:  logger,
	}
}

// Generate takes a path to scan and a list of SBOM mediatypes (with an
// optional version for CycloneDX and SPDX), and invokes the syft CLI scan
// command. The CLI is instructed to write the SBOM to
// <layers>/<layer>.sbom.<ext> as defined by the buildpack spec. Additionally,
// CycloneDX & SPDX outputs are modified to make the output reproducible
// (Paketo RFCs 38 & 49).
func (s SyftCLIScanner) GenerateSBOM(scanPath, layersPath, layerName string, mediaTypes ...string) error {
	sbomWritePaths := make(map[string]string)
	args := []string{"scan", "--quiet"}

	s.logger.Debug.Process("Generating SBOM")
	s.logger.Debug.Subprocess("Generating syft CLI args from provided mediatypes %s", mediaTypes)
	for _, mediatype := range mediaTypes {
		syftOutputFormat, err := s.specMediatypeToSyftOutputFormat(mediatype)
		if err != nil {
			return fmt.Errorf("failed to convert mediatype %s to syft output format: %w", mediatype, err)
		}

		extension, err := Format(mediatype).Extension()
		if err != nil {
			return err
		}

		// Layer SBOM write location during build is <layers>/<layer>.sbom.<ext> (CNB spec)
		sbomWritePaths[mediatype] = filepath.Join(layersPath, fmt.Sprintf("%s.sbom.%s", layerName, extension))
		args = append(args, "--output", fmt.Sprintf("%s=%s", syftOutputFormat, sbomWritePaths[mediatype]))
	}

	args = append(args, scanPath)

	s.logger.Debug.Subprocess("Executing syft CLI with args %v", args)
	if err := s.syftCLI.Execute(pexec.Execution{
		Args:   args,
		Stdout: s.logger.ActionWriter,
		Stderr: s.logger.ActionWriter,
	}); err != nil {
		return fmt.Errorf("failed to execute syft cli with args '%s': %w.\nYou might be missing a buildpack that provides the syft CLI", args, err)
	}

	// Make SBOM outputs reproducible
	for _, mediatype := range mediaTypes {
		if strings.HasPrefix(mediatype, CycloneDXFormat) {
			s.logger.Debug.Subprocess("Processing syft CLI CycloneDX SBOM output to make it reproducible")
			err := s.makeCycloneDXReproducible(sbomWritePaths[mediatype])
			if err != nil {
				return fmt.Errorf("failed to make CycloneDX SBOM reproducible: %w", err)
			}
		} else if strings.HasPrefix(mediatype, SPDXFormat) {
			s.logger.Debug.Subprocess("Processing syft CLI SPDX SBOM output to make it reproducible")
			err := s.makeSPDXReproducible(sbomWritePaths[mediatype])
			if err != nil {
				return fmt.Errorf("failed to make SPDX SBOM reproducible: %w", err)
			}
		}
	}

	s.logger.Debug.Break()
	return nil
}

// This method takes an SBOM mediatype name as defined by the buildpack spec,
// (with an optional version param for CycloneDX and SPDX, e.g.
// "application/vnd.cyclonedx+json;version=1.4") and returns the output format
// understood by syft tooling (e.g. "cyclonedx-json@1.4").
// Refer github.com/anchore/syft/blob/v1.11.1/cmd/syft/internal/options/writer.go#L86
func (s SyftCLIScanner) specMediatypeToSyftOutputFormat(mediatype string) (string, error) {
	optionalVersionParam, err := Format(mediatype).VersionParam()
	if err != nil {
		return "", err
	}
	if optionalVersionParam != "" {
		optionalVersionParam = "@" + optionalVersionParam
	}

	switch {
	case strings.HasPrefix(mediatype, CycloneDXFormat):
		return "cyclonedx-json" + optionalVersionParam, nil
	case strings.HasPrefix(mediatype, SPDXFormat):
		return "spdx-json" + optionalVersionParam, nil
	case strings.HasPrefix(mediatype, SyftFormat):
		// The syft tool does not support providing a version for the syft mediatype.
		if optionalVersionParam != "" {
			return "", fmt.Errorf("The syft mediatype does not allow providing a ;version=<ver> param. Got: %s", mediatype)
		}
		return "syft-json", nil
	default:
		return "", fmt.Errorf("mediatype %s matched none of the known mediatypes. Valid values are %s, with an optional version param for CycloneDX and SPDX", mediatype, []string{CycloneDXFormat, SPDXFormat, SyftFormat})
	}
}

// Makes CycloneDX SBOM more reproducible.
// Remove fields serialNumber and metadata.timestamp.
// See https://github.com/paketo-buildpacks/rfcs/blob/main/text/0038-cdx-syft-sbom.md#amendment-sbom-reproducibility
func (s SyftCLIScanner) makeCycloneDXReproducible(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to read CycloneDX JSON file %s:%w", path, err)
	}
	defer in.Close()

	input := map[string]interface{}{}
	if err := json.NewDecoder(in).Decode(&input); err != nil {
		return fmt.Errorf("unable to decode CycloneDX JSON %s: %w", path, err)
	}

	delete(input, "serialNumber")

	if md, exists := input["metadata"]; exists {
		if metadata, ok := md.(map[string]interface{}); ok {
			delete(metadata, "timestamp")
		}
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to open CycloneDX JSON for writing %s: %w", path, err)
	}
	defer out.Close()

	if err := json.NewEncoder(out).Encode(input); err != nil {
		return fmt.Errorf("unable to encode CycloneDX: %w", err)
	}

	return nil
}

// Makes SPDX SBOM more reproducible.
// Ensure documentNamespace and creationInfo.created have reproducible values.
// The method respects $SOURCE_DATE_EPOCH for created timestamp if set.
// See github.com/paketo-buildpacks/rfcs/blob/main/text/0049-reproducible-spdx.md
func (s SyftCLIScanner) makeSPDXReproducible(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to read SPDX JSON file %s:%w", path, err)
	}
	defer in.Close()

	input := map[string]interface{}{}
	if err := json.NewDecoder(in).Decode(&input); err != nil {
		return fmt.Errorf("unable to decode SPDX JSON %s: %w", path, err)
	}

	// Makes the creationInfo reproducible so a hash can be taken for the
	// documentNamespace
	if creationInfo, ok := input["creationInfo"].(map[string]interface{}); ok {
		creationInfo["created"] = time.Time{} // This is the zero-valued time

		sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH")
		if sourceDateEpoch != "" {
			sde, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse SOURCE_DATE_EPOCH: %w", err)
			}
			creationInfo["created"] = time.Unix(sde, 0).UTC()
		}
		input["creationInfo"] = creationInfo
	}

	if namespace, ok := input["documentNamespace"].(string); ok {
		delete(input, "documentNamespace")

		data, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("failed to checksum SPDX document: %w", err)
		}

		uri, err := url.Parse(namespace)
		if err != nil {
			return fmt.Errorf("failed to parse SPDX documentNamespace url: %w", err)
		}

		uri.Host = "paketo.io"
		uri.Path = strings.Replace(uri.Path, "syft", "packit", 1)
		oldBase := filepath.Base(uri.Path)
		source, _, _ := strings.Cut(oldBase, "-")
		newBase := fmt.Sprintf("%s-%s", source, uuid.NewSHA1(uuid.NameSpaceURL, data))
		uri.Path = strings.Replace(uri.Path, oldBase, newBase, 1)

		input["documentNamespace"] = uri.String()
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to open SPDX JSON for writing %s: %w", path, err)
	}
	defer out.Close()

	if err := json.NewEncoder(out).Encode(input); err != nil {
		return fmt.Errorf("unable to encode SPDX: %w", err)
	}
	return nil
}
