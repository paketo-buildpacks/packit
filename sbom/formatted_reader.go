package sbom

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/sbom"
	"github.com/google/uuid"
	"github.com/mitchellh/hashstructure/v2"
)

// FormattedReader outputs the SBoM in a specified format.
type FormattedReader struct {
	m           sync.Mutex
	sbom        SBOM
	rawFormatID string
	format      sbom.Format
	reader      io.Reader
}

// NewFormattedReader creates an instance of FormattedReader given an SBOM and
// Format.
func NewFormattedReader(s SBOM, f Format) *FormattedReader {
	// For backward compatibility, caller can pass f as a format ID like
	// "cyclonedx-1.3-json" or as a media type like
	// 'application/vnd.cyclonedx+json'
	sbomFormat, err := sbomFormatByID(sbom.FormatID(f))
	if err != nil {
		sbomFormat, err = sbomFormatByMediaType(string(f))
		if err != nil {
			// Defer throwing an error until Read() is called
			return &FormattedReader{sbom: s, rawFormatID: string(f), format: nil}
		}
	}
	return &FormattedReader{sbom: s, rawFormatID: string(sbomFormat.ID()), format: sbomFormat}
}

// Read implements the io.Reader interface to output the contents of the
// formatted SBoM.
func (f *FormattedReader) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		if f.format == nil {
			return 0, fmt.Errorf("failed to format sbom: '%s' is not a valid SBOM format identifier", f.rawFormatID)
		}

		output, err := syft.Encode(f.sbom.syft, f.format)
		if err != nil {
			// not tested
			return 0, fmt.Errorf("failed to format sbom: %w", err)
		}

		// Makes CycloneDX SBOM more reproducible, see
		// https://github.com/paketo-buildpacks/packit/issues/367 for more details.
		if f.format.ID() == "cyclonedx-1.3-json" || f.format.ID() == "cyclonedx-1-json" {
			var cycloneDXOutput map[string]interface{}
			err = json.Unmarshal(output, &cycloneDXOutput)
			if err != nil {
				return 0, fmt.Errorf("failed to modify CycloneDX SBOM for reproducibility: %w", err)
			}

			if metadata, ok := cycloneDXOutput["metadata"].(map[string]interface{}); ok {
				delete(metadata, "timestamp")
				cycloneDXOutput["metadata"] = metadata
			}

			delete(cycloneDXOutput, "serialNumber")

			output, err = json.Marshal(cycloneDXOutput)
			if err != nil {
				return 0, fmt.Errorf("failed to modify CycloneDX SBOM for reproducibility: %w", err)
			}
		}

		// Makes SPDX SBOM more reproducible, see
		// https://github.com/paketo-buildpacks/packit/issues/368 for more details.
		if f.format.ID() == "spdx-2-json" {
			hash, err := hashstructure.Hash(f.sbom.syft, hashstructure.FormatV2, nil)
			if err != nil {
				// not tested
				return 0, fmt.Errorf("failed to modify SPDX SBOM for reproducibility: %w", err)
			}
			hashBytes := make([]byte, binary.MaxVarintLen64)
			binary.PutUvarint(hashBytes, hash)

			var spdxOutput map[string]interface{}

			err = json.Unmarshal(output, &spdxOutput)
			if err != nil {
				return 0, fmt.Errorf("failed to modify SPDX SBOM for reproducibility: %w", err)
			}
			if namespace, ok := spdxOutput["documentNamespace"].(string); ok {
				uri, err := url.Parse(namespace)
				if err != nil {
					// not tested
					return 0, err
				}

				uri.Host = "paketo.io"
				uri.Path = strings.Replace(uri.Path, "syft", "packit", 1)
				oldBase := filepath.Base(uri.Path)
				source, _, _ := strings.Cut(oldBase, "-")
				newBase := fmt.Sprintf("%s-%s", source, uuid.NewSHA1(uuid.NameSpaceURL, hashBytes))
				uri.Path = strings.Replace(uri.Path, oldBase, newBase, 1)

				spdxOutput["documentNamespace"] = uri.String()
			}

			if creationInfo, ok := spdxOutput["creationInfo"].(map[string]interface{}); ok {
				creationInfo["created"] = time.Time{} // This is the zero-valued time

				source_date_epoch := os.Getenv("SOURCE_DATE_EPOCH")
				if source_date_epoch != "" {
					sde, err := strconv.ParseInt(source_date_epoch, 10, 64)
					if err != nil {
						return 0, fmt.Errorf("failed to parse SOURCE_DATE_EPOCH: %w", err)
					}
					creationInfo["created"] = time.Unix(sde, 0).UTC()
				}
				spdxOutput["creationInfo"] = creationInfo
			}

			output, err = json.Marshal(spdxOutput)
			if err != nil {
				return 0, fmt.Errorf("failed to modify SPDX SBOM for reproducibility: %w", err)
			}
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}
