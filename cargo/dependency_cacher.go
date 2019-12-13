package cargo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:generate faux --interface Downloader --output fakes/downloader.go
type Downloader interface {
	Drop(root, uri string) (io.ReadCloser, error)
}

type DependencyCacher struct {
	downloader Downloader
}

func NewDependencyCacher(downloader Downloader) DependencyCacher {
	return DependencyCacher{
		downloader: downloader,
	}
}

func (dc DependencyCacher) Cache(root string, deps []ConfigMetadataDependency) ([]ConfigMetadataDependency, error) {
	dir := filepath.Join(root, "dependencies")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create dependencies directory: %s", err)
	}

	var dependencies []ConfigMetadataDependency
	for _, dep := range deps {
		source, err := dc.downloader.Drop("", dep.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to download dependency: %s", err)
		}

		validatedSource := NewValidatedReader(source, dep.SHA256)

		destination, err := os.Create(filepath.Join(dir, dep.SHA256))
		if err != nil {
			return nil, fmt.Errorf("failed to create destination file: %s", err)
		}

		_, err = io.Copy(destination, validatedSource)
		if err != nil {
			return nil, fmt.Errorf("failed to copy dependency: %s", err)
		}

		err = destination.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close dependency destination: %s", err)
		}

		err = source.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close dependency source: %s", err)
		}

		dep.URI = fmt.Sprintf("file:///dependencies/%s", dep.SHA256)
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}
