package cyclonedxhelpers

import (
	"testing"

	// "github.com/CycloneDX/cyclonedx-go"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
	"github.com/stretchr/testify/assert"

	"github.com/anchore/syft/syft/file"
	"github.com/anchore/syft/syft/pkg"
)

func Test_encodeComponentProperties(t *testing.T) {
	epoch := 2
	tests := []struct {
		name     string
		input    pkg.Package
		expected *[]cyclonedx.Property
	}{
		{
			name:     "no metadata",
			input:    pkg.Package{},
			expected: nil,
		},
		{
			name: "from apk",
			input: pkg.Package{
				FoundBy: "cataloger",
				Locations: file.NewLocationSet(
					file.NewLocationFromCoordinates(file.Coordinates{RealPath: "test"}),
				),
				Metadata: pkg.ApkMetadata{
					Package:       "libc-utils",
					OriginPackage: "libc-dev",
					Maintainer:    "Natanael Copa <ncopa@alpinelinux.org>",
					Version:       "0.7.2-r0",
					Architecture:  "x86_64",
					URL:           "http://alpinelinux.org",
					Description:   "Meta package to pull in correct libc",
					Size:          0,
					InstalledSize: 4096,
					Dependencies:  []string{"musl-utils"},
					Provides:      []string{"so:libc.so.1"},
					Checksum:      "Q1p78yvTLG094tHE1+dToJGbmYzQE=",
					GitCommit:     "97b1c2842faa3bfa30f5811ffbf16d5ff9f1a479",
					Files:         []pkg.ApkFileRecord{},
				},
			},
			expected: &[]cyclonedx.Property{
				{Name: "syft:package:foundBy", Value: "cataloger"},
				{Name: "syft:location:0:path", Value: "test"},
				{Name: "syft:metadata:gitCommitOfApkPort", Value: "97b1c2842faa3bfa30f5811ffbf16d5ff9f1a479"},
				{Name: "syft:metadata:installedSize", Value: "4096"},
				{Name: "syft:metadata:originPackage", Value: "libc-dev"},
				{Name: "syft:metadata:provides:0", Value: "so:libc.so.1"},
				{Name: "syft:metadata:pullChecksum", Value: "Q1p78yvTLG094tHE1+dToJGbmYzQE="},
				{Name: "syft:metadata:pullDependencies:0", Value: "musl-utils"},
				{Name: "syft:metadata:size", Value: "0"},
			},
		},
		{
			name: "from dpkg",
			input: pkg.Package{
				MetadataType: pkg.DpkgMetadataType,
				Metadata: pkg.DpkgMetadata{
					Package:       "tzdata",
					Version:       "2020a-0+deb10u1",
					Source:        "tzdata-dev",
					SourceVersion: "1.0",
					Architecture:  "all",
					InstalledSize: 3036,
					Maintainer:    "GNU Libc Maintainers <debian-glibc@lists.debian.org>",
					Files:         []pkg.DpkgFileRecord{},
				},
			},
			expected: &[]cyclonedx.Property{
				{Name: "syft:package:metadataType", Value: "DpkgMetadata"},
				{Name: "syft:metadata:installedSize", Value: "3036"},
				{Name: "syft:metadata:source", Value: "tzdata-dev"},
				{Name: "syft:metadata:sourceVersion", Value: "1.0"},
			},
		},
		{
			name: "from go bin",
			input: pkg.Package{
				Name:         "golang.org/x/net",
				Version:      "v0.0.0-20211006190231-62292e806868",
				Language:     pkg.Go,
				Type:         pkg.GoModulePkg,
				MetadataType: pkg.GolangBinMetadataType,
				Metadata: pkg.GolangBinMetadata{
					GoCompiledVersion: "1.17",
					Architecture:      "amd64",
					H1Digest:          "h1:KlOXYy8wQWTUJYFgkUI40Lzr06ofg5IRXUK5C7qZt1k=",
				},
			},
			expected: &[]cyclonedx.Property{
				{Name: "syft:package:language", Value: pkg.Go.String()},
				{Name: "syft:package:metadataType", Value: "GolangBinMetadata"},
				{Name: "syft:package:type", Value: "go-module"},
				{Name: "syft:metadata:architecture", Value: "amd64"},
				{Name: "syft:metadata:goCompiledVersion", Value: "1.17"},
				{Name: "syft:metadata:h1Digest", Value: "h1:KlOXYy8wQWTUJYFgkUI40Lzr06ofg5IRXUK5C7qZt1k="},
			},
		},
		{
			name: "from rpm",
			input: pkg.Package{
				Name:         "dive",
				Version:      "0.9.2-1",
				Type:         pkg.RpmPkg,
				MetadataType: pkg.RpmMetadataType,
				Metadata: pkg.RpmMetadata{
					Name:      "dive",
					Epoch:     &epoch,
					Arch:      "x86_64",
					Release:   "1",
					Version:   "0.9.2",
					SourceRpm: "dive-0.9.2-1.src.rpm",
					Size:      12406784,
					Vendor:    "",
					Files:     []pkg.RpmdbFileRecord{},
				},
			},
			expected: &[]cyclonedx.Property{
				{Name: "syft:package:metadataType", Value: "RpmMetadata"},
				{Name: "syft:package:type", Value: "rpm"},
				{Name: "syft:metadata:epoch", Value: "2"},
				{Name: "syft:metadata:release", Value: "1"},
				{Name: "syft:metadata:size", Value: "12406784"},
				{Name: "syft:metadata:sourceRpm", Value: "dive-0.9.2-1.src.rpm"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := encodeComponent(test.input)
			assert.Equal(t, test.expected, c.Properties)
		})
	}
}
