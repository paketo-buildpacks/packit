package cargo

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cloudfoundry/packit/scribe"
)

type File struct {
	io.ReadCloser

	Name string
	Info os.FileInfo
}

type FileInfo struct {
	name  string
	size  int
	mode  uint32
	mtime time.Time
}

func NewFileInfo(name string, size int, mode uint32, mtime time.Time) FileInfo {
	return FileInfo{
		name:  name,
		size:  size,
		mode:  mode,
		mtime: mtime,
	}
}

func (fi FileInfo) Name() string {
	return fi.name
}

func (fi FileInfo) Size() int64 {
	return int64(fi.size)
}

func (fi FileInfo) Mode() os.FileMode {
	return os.FileMode(fi.mode)
}

func (fi FileInfo) ModTime() time.Time {
	return fi.mtime
}

func (fi FileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi FileInfo) Sys() interface{} {
	return nil
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
		hdr, err := tar.FileInfoHeader(file.Info, file.Name)
		if err != nil {
			return fmt.Errorf("failed to create header for file %q: %w", file.Name, err)
		}

		hdr.Name = file.Name

		err = tw.WriteHeader(hdr)
		if err != nil {
			return fmt.Errorf("failed to write header to tarball: %w", err)
		}

		_, err = io.Copy(tw, file)
		if err != nil {
			return fmt.Errorf("failed to write file to tarball: %w", err)
		}

		file.Close()
	}

	b.logger.Break()

	return nil
}
