package internal

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/cargo"
)

type BuildpackInspector struct{}

func NewBuildpackInspector() BuildpackInspector {
	return BuildpackInspector{}
}

func (i BuildpackInspector) Dependencies(path string) ([]cargo.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	indexJSON, err := fetchArchivedFile(tar.NewReader(file), "index.json")
	if err != nil {
		return nil, err
	}

	var index struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}

	err = json.NewDecoder(indexJSON).Decode(&index)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	manifest, err := fetchArchivedFile(tar.NewReader(file), filepath.Join("blobs", "sha256", strings.TrimPrefix(index.Manifests[0].Digest, "sha256:")))
	if err != nil {
		return nil, err
	}

	buildpackageDigest := index.Manifests[0].Digest

	var m struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}

	err = json.NewDecoder(manifest).Decode(&m)
	if err != nil {
		return nil, err
	}

	var configs []cargo.Config
	for _, layer := range m.Layers {
		_, err = file.Seek(0, 0)
		if err != nil {
			return nil, err
		}

		buildpack, err := fetchArchivedFile(tar.NewReader(file), filepath.Join("blobs", "sha256", strings.TrimPrefix(layer.Digest, "sha256:")))
		if err != nil {
			return nil, err
		}

		buildpackGR, err := gzip.NewReader(buildpack)
		if err != nil {
			return nil, fmt.Errorf("failed to read buildpack gzip: %w", err)
		}
		defer buildpackGR.Close()

		buildpackTOML, err := fetchArchivedFile(tar.NewReader(buildpackGR), "buildpack.toml")
		if err != nil {
			return nil, err
		}

		var config cargo.Config
		err = cargo.DecodeConfig(buildpackTOML, &config)
		if err != nil {
			return nil, err
		}

		if len(config.Order) > 0 {
			config.Buildpack.SHA256 = buildpackageDigest
		}
		configs = append(configs, config)
	}

	if len(configs) == 1 {
		configs[0].Buildpack.SHA256 = buildpackageDigest
	}

	return configs, nil
}

func fetchArchivedFile(tr *tar.Reader, filename string) (io.Reader, error) {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(hdr.Name, filename) {
			return tr, nil
		}
	}

	return nil, fmt.Errorf("failed to fetch archived file %s", filename)
}
