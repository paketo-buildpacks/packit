package reload

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
)

var LiveReloadEnabledEnvVar = "BP_LIVE_RELOAD_ENABLED"
var WatchExecRequirement = packit.BuildPlanRequirement{
	Name: "watchexec",
	Metadata: map[string]interface{}{
		"launch": true,
	}}

func ShouldEnableLiveReload() (bool, error) {
	if reload, ok := os.LookupEnv(LiveReloadEnabledEnvVar); ok {
		if shouldEnableReload, err := strconv.ParseBool(reload); err != nil {
			return false, fmt.Errorf("failed to parse %s value %s: %w", LiveReloadEnabledEnvVar, reload, err)
		} else if shouldEnableReload {
			return true, nil
		}
	}
	return false, nil
}

// ReloadableProcessSpec contains information to be translated directly into watchexec arguments
type ReloadableProcessSpec struct {
	// WatchPaths will translate into --watch args
	// Optional. If len == 0, then no --watch args will be provided
	WatchPaths []string

	// IgnorePaths will translate into --ignore args
	// Optional. If len == 0, then no --ignore args will be provided
	IgnorePaths []string

	// Shell will translate into --shell
	// Optional. If not provided, will use "none"
	Shell string

	// VerbosityLevel will translate into -v
	// If VerbosityLevel is 0, no -v arg will be provided.
	// If VerbosityLevel is greater than 0, the appropriate number of v's will be added
	// E.g. VerbosityLevel = 2 => -vv
	// E.g. VerbosityLevel = 4 => -vvvv
	VerbosityLevel int
}

func TransformReloadableProcesses(originalProcess packit.Process, spec ReloadableProcessSpec) (nonReloadable packit.Process, reloadable packit.Process) {
	nonReloadable = originalProcess
	nonReloadable.Default = false

	reloadable = originalProcess
	reloadable.Type = fmt.Sprintf("reload-%s", originalProcess.Type)
	reloadable.Command = "watchexec"
	reloadable.Args = buildArgs(originalProcess, spec)

	return nonReloadable, reloadable
}

func buildArgs(originalProcess packit.Process, spec ReloadableProcessSpec) []string {
	args := []string{
		"--restart",
	}

	for _, watchPath := range spec.WatchPaths {
		args = append(args, "--watch", watchPath)
	}

	for _, ignorePath := range spec.IgnorePaths {
		args = append(args, "--ignore", ignorePath)
	}

	shell := "none"
	if spec.Shell != "" {
		shell = spec.Shell
	}
	args = append(args, "--shell", shell)

	if spec.VerbosityLevel > 0 {
		args = append(args, "-"+strings.Repeat("v", spec.VerbosityLevel))
	}

	args = append(args,
		"--",
		originalProcess.Command)
	args = append(args, originalProcess.Args...)
	return args
}
