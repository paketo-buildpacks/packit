package cargo

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/cloudfoundry/packit/scribe"
)

type File struct {
	io.ReadCloser

	Name string
	Size int64
	Mode int64
}

type TarBuilder struct {
	logger scribe.Logger
}

func NewTarBuilder(logger scribe.Logger) TarBuilder {
	return TarBuilder{
		logger: logger,
	}
}

func (b TarBuilder) Build(path string, files []File) error {
	b.logger.Process("Building tarball: %s", path)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create tarball: %s", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		b.logger.Subprocess(file.Name)
		err = tw.WriteHeader(&tar.Header{
			Name: file.Name,
			Size: file.Size,
			Mode: file.Mode,
		})
		if err != nil {
			return fmt.Errorf("failed to write header to tarball: %s", err)
		}

		_, err = io.Copy(tw, file)
		if err != nil {
			return fmt.Errorf("failed to write file to tarball: %s", err)
		}

		file.Close()
	}

	b.logger.Break()

	return nil
}
