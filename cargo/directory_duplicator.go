package cargo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type DirectoryDuplicator struct{}

func NewDirectoryDuplicator() DirectoryDuplicator {
	return DirectoryDuplicator{}
}

func (d DirectoryDuplicator) Duplicate(source, sink string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("source dir does not exist: %s", err)
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %s", err)
		}

		destPath := filepath.Join(sink, relPath)
		if info.IsDir() {
			err := os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return fmt.Errorf("duplicate error creating dir: %s", err)
			}
		} else if os.ModeType&info.Mode() == 0 {
			src, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening source file failed: %s", err)
			}
			defer src.Close()

			srcInfo, err := src.Stat()
			if err != nil {
				return fmt.Errorf("unable to stat source file: %s", err)
			}

			dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, srcInfo.Mode())
			if err != nil {
				return fmt.Errorf("duplicate error creating file: %s", err)
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return fmt.Errorf("copy dst to src failed: %s", err)
			}

		}
		return nil
	})
}
