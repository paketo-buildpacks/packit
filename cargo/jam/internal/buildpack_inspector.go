package internal

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit/cargo"
)

type BuildpackInspector struct{}

func NewBuildpackInspector() BuildpackInspector {
	return BuildpackInspector{}
}

func (i BuildpackInspector) Dependencies(path string) ([]cargo.ConfigMetadataDependency, map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		if filepath.Base(hdr.Name) == "buildpack.toml" {
			var config cargo.Config
			err = cargo.DecodeConfig(tr, &config)
			if err != nil {
				return nil, nil, err
			}
			return config.Metadata.Dependencies, config.Metadata.DefaultVersions, nil
		}
	}

	return nil, nil, errors.New("failed to find buildpack.toml in buildpack tarball")
}
