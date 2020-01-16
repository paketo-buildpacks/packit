package fs

import (
	"io/ioutil"
)

func IsEmptyDir(path string) bool {
	contents, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}

	return len(contents) == 0
}
