package fs

import (
	"io/ioutil"
)

// IsEmptyDir checks to see if a directory exists and is empty.
func IsEmptyDir(path string) bool {
	contents, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}

	return len(contents) == 0
}
