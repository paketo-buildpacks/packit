package packit

// Platform contains the context of the buildpack platform including its
// location on the filesystem.
type Platform struct {
	// Path provides the location of the platform directory on the filesystem.
	// This location can be used to find platform extensions like service
	// bindings.
	Path string
}
