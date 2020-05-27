package cargo

import "github.com/paketo-buildpacks/packit/fs"

type DirectoryDuplicator struct{}

func NewDirectoryDuplicator() DirectoryDuplicator {
	return DirectoryDuplicator{}
}

func (d DirectoryDuplicator) Duplicate(source, destination string) error {
	return fs.Copy(source, destination)
}
