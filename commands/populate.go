package commands

import (
	"fmt"
	. "github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"math"
	"os"
	"path/filepath"
)

// PopulateExecDCommands will populate the named exec.d commands into a new layer under the "exec.d" folder.
// The commands will be executed in `alphabetically ascending order by file name` as per the buildpacks launch mechanism.
// See https://github.com/buildpacks/spec/blob/main/buildpack.md#launch.
// Commands given to this function will be given a numerical prefix so that they will be executed in the slice order.
// E.g. PopulateExecDCommands(context, "layerName", "helper", "command", "other-process") will result in:
// - <layers>/layerName/exec.d/0-helper
// - <layers>/layerName/exec.d/1-command
// - <layers>/layerName/exec.d/2-other-process
// Do not forget to add these files to the `buildpack.toml` so they are available.
// If enough commands are given, this function will pad the prefix with 0 so that they are in the correct alphabetical order.
// E.g. PopulateExecDCommands(context, "layerName", "cmd0", ... , "cmd10", ... "cmd100") will result in:
// - <layers>/layerName/exec.d/000-cmd0
// - <layers>/layerName/exec.d/010-cmd10
// - <layers>/layerName/exec.d/100-cmd100
func PopulateExecDCommands(context BuildContext, layerName string, commands ...string) (Layer, error) {
	layer, err := context.Layers.Get(layerName)
	if err != nil {
		return layer, err
	}

	layer.Launch = true

	if len(commands) < 1 {
		return layer, nil
	}

	execdDir := filepath.Join(layer.Path, "exec.d")
	err = os.MkdirAll(execdDir, os.ModePerm)
	if err != nil {
		return layer, err
	}

	lexicalWidth := 1 + int(math.Log10(float64(len(commands))))

	for i, command := range commands {
		filename := filepath.Join("bin", command)
		source := filepath.Join(context.CNBPath, filename)
		destination := filepath.Join(execdDir, fmt.Sprintf("%0*d-%s", lexicalWidth, i, command))

		if exists, err := fs.Exists(source); !exists {
			return layer, fmt.Errorf("file %s does not exist. Be sure to include it in the buildpack.toml", filename)
		} else if err != nil {
			return layer, err
		}

		err = fs.Copy(source, destination)
		if err != nil {
			return layer, err
		}
	}

	return layer, nil
}
