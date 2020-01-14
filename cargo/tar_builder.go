package cargo

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cloudfoundry/packit/scribe"
)

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

	directories := map[string]struct{}{}
	for _, file := range files {
		path := filepath.Dir(file.Name)
		for path != "." {
			directories[path] = struct{}{}

			path = filepath.Dir(path)
		}
	}

	for dir := range directories {
		files = append(files, File{
			Name: dir,
			Info: NewFileInfo(filepath.Base(dir), 0, os.ModePerm|os.ModeDir, time.Now()),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	for _, file := range files {
		b.logger.Subprocess(file.Name)
		hdr, err := tar.FileInfoHeader(file.Info, file.Name)
		if err != nil {
			return fmt.Errorf("failed to create header for file %q: %w", file.Name, err)
		}

		hdr.Name = file.Name

		err = tw.WriteHeader(hdr)
		if err != nil {
			return fmt.Errorf("failed to write header to tarball: %w", err)
		}

		if file.ReadCloser != nil {
			_, err = io.Copy(tw, file)
			if err != nil {
				return fmt.Errorf("failed to write file to tarball: %w", err)
			}

			file.Close()
		}
	}

	b.logger.Break()

	return nil
}
