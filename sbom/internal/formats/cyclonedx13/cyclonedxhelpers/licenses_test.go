package cyclonedxhelpers

import (
	"testing"

	"github.com/anchore/syft/syft/pkg"
	// "github.com/CycloneDX/cyclonedx-go"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
	"github.com/stretchr/testify/assert"
)

func Test_encodeLicense(t *testing.T) {
	tests := []struct {
		name     string
		input    pkg.Package
		expected *cyclonedx.Licenses
	}{
		{
			name:     "no licenses",
			input:    pkg.Package{},
			expected: nil,
		},
		{
			name: "no SPDX licenses",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicense("RandomLicense"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					License: &cyclonedx.License{
						Name: "RandomLicense",
					},
				},
			},
		},
		{
			name: "single SPDX ID and Non SPDX ID",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicense("mit"),
					pkg.NewLicense("FOOBAR"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					License: &cyclonedx.License{
						Name: "FOOBAR",
					},
				},
				{
					License: &cyclonedx.License{
						ID: "MIT",
					},
				},
			},
		},
		{
			name: "with complex SPDX license expression",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicense("MIT AND GPL-3.0-only"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					Expression: "MIT AND GPL-3.0-only",
				},
			},
		},
		{
			name: "with multiple complex SPDX license expression",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicense("MIT AND GPL-3.0-only"),
					pkg.NewLicense("MIT AND GPL-3.0-only WITH Classpath-exception-2.0"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					Expression: "(MIT AND GPL-3.0-only) AND (MIT AND GPL-3.0-only WITH Classpath-exception-2.0)",
				},
			},
		},
		{
			name: "with multiple URLs and expressions",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicenseFromURLs("MIT", "https://opensource.org/licenses/MIT", "https://spdx.org/licenses/MIT.html"),
					pkg.NewLicense("MIT AND GPL-3.0-only"),
					pkg.NewLicenseFromURLs("FakeLicense", "htts://someurl.com"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					License: &cyclonedx.License{
						Name: "FakeLicense",
						URL:  "htts://someurl.com",
					},
				},
				{
					License: &cyclonedx.License{
						Name: "MIT AND GPL-3.0-only",
					},
				},
				{
					License: &cyclonedx.License{
						ID:  "MIT",
						URL: "https://opensource.org/licenses/MIT",
					},
				},
				{
					License: &cyclonedx.License{
						ID:  "MIT",
						URL: "https://spdx.org/licenses/MIT.html",
					},
				},
			},
		},
		{
			name: "with multiple values licenses are deduplicated",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicense("Apache-2"),
					pkg.NewLicense("Apache-2.0"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					License: &cyclonedx.License{
						ID: "Apache-2.0",
					},
				},
			},
		},
		{
			name: "with multiple URLs and single with no URLs",
			input: pkg.Package{
				Licenses: pkg.NewLicenseSet(
					pkg.NewLicense("MIT"),
					pkg.NewLicenseFromURLs("MIT", "https://opensource.org/licenses/MIT", "https://spdx.org/licenses/MIT.html"),
					pkg.NewLicense("MIT AND GPL-3.0-only"),
				),
			},
			expected: &cyclonedx.Licenses{
				{
					License: &cyclonedx.License{
						ID:  "MIT",
						URL: "https://opensource.org/licenses/MIT",
					},
				},
				{
					License: &cyclonedx.License{
						ID:  "MIT",
						URL: "https://spdx.org/licenses/MIT.html",
					},
				},
				{
					License: &cyclonedx.License{
						Name: "MIT AND GPL-3.0-only",
					},
				},
			},
		},
		// TODO: do we drop the non SPDX ID license and do a single expression
		// OR do we keep the non SPDX ID license and do multiple licenses where the complex
		// expressions are set as the NAME field?
		//{
		//	name: "with multiple complex SPDX license expression and a non spdx id",
		//	input: pkg.Package{
		//		Licenses: []pkg.License{
		//			{
		//				SPDXExpression: "MIT AND GPL-3.0-only",
		//			},
		//			{
		//				SPDXExpression: "MIT AND GPL-3.0-only WITH Classpath-exception-2.0",
		//			},
		//			{
		//				Value: "FOOBAR",
		//			},
		//		},
		//	},
		//	expected: &cyclonedx.Licenses{
		//		{
		//			Expression: "(MIT AND GPL-3.0-only) AND (MIT AND GPL-3.0-only WITH Classpath-exception-2.0)",
		//		},
		//	},
		//},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, encodeLicenses(test.input))
		})
	}
}
