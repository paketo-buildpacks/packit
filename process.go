package packit

// Process represents a process to be run during the launch phase as described
// in the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#launch. The
// fields of the process are describe in the specification of the launch.toml
// file:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#launchtoml-toml.
type Process struct {
	// Type is an identifier to describe the type of process to be executed, eg.
	// "web".
	Type string `toml:"type"`

	// Command is the start command to be executed at launch.
	Command string `toml:"command"`

	// Args is a list of arguments to be passed to the command at launch.
	Args []string `toml:"args"`

	// Direct indicates whether the process should bypass the shell when invoked.
	Direct bool `toml:"direct"`

	// Default indicates if this process should be the default when launched.
	Default bool `toml:"default,omitempty"`

	// WorkingDirectory indicates if this process should be run in a working
	// directory other than the application directory. This can either be an
	// absolute path or one relative to the default application directory.
	WorkingDirectory string `toml:"working-directory,omitempty"`
}
