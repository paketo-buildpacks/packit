package cyclonedxhelpers

import (
	"fmt"
	"strings"

	"github.com/anchore/syft/syft/pkg"
)

// We must copy this helper in because it's not exported from
// syft/formats/common/cyclonedxhelpers
func encodeAuthor(p pkg.Package) string {
	if hasMetadata(p) {
		switch metadata := p.Metadata.(type) {
		case pkg.NpmPackageJSONMetadata:
			return metadata.Author
		case pkg.PythonPackageMetadata:
			author := metadata.Author
			if metadata.AuthorEmail != "" {
				if author == "" {
					return metadata.AuthorEmail
				}
				author += fmt.Sprintf(" <%s>", metadata.AuthorEmail)
			}
			return author
		case pkg.GemMetadata:
			if len(metadata.Authors) > 0 {
				return strings.Join(metadata.Authors, ",")
			}
			return ""
		}
	}
	return ""
}
