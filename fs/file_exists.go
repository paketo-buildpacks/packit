package fs

import (
	"errors"
	"os"
)

// FileExists returns true if a file or directory at the given path is present and false otherwise.
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
