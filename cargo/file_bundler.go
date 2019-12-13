package cargo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileBundler struct{}

func NewFileBundler() FileBundler {
	return FileBundler{}
}

func (b FileBundler) Bundle(root string, paths []string, config Config) ([]File, error) {
	var files []File
	for _, path := range paths {
		file := File{Name: path}

		switch path {
		case "buildpack.toml":
			buf := bytes.NewBuffer(nil)
			err := EncodeConfig(buf, config)
			if err != nil {
				return nil, fmt.Errorf("error encoding buildpack.toml: %s", err)
			}

			file.ReadCloser = ioutil.NopCloser(buf)
			file.Size = int64(buf.Len())
			file.Mode = int64(0644)

		default:
			fd, err := os.Open(filepath.Join(root, path))
			if err != nil {
				return nil, fmt.Errorf("error opening included file: %s", err)
			}

			info, err := fd.Stat()
			if err != nil {
				return nil, fmt.Errorf("error stating included file: %s", err)
			}

			file.ReadCloser = fd
			file.Size = info.Size()
			file.Mode = int64(info.Mode())

		}

		files = append(files, file)
	}

	return files, nil
}
