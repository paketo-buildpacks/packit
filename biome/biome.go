// Package biome provide useful enivonment variable parsing functionality
package biome

import (
	"fmt"
	"os"
	"strconv"
)

// GetBool takes the name of an environment variable as an argument and returns
// its boolean value if available or an error if it is unable to parse the
// value from the environment variable into a bool.
//
// If the given environment variable is unset then the function will return
// false.
func GetBool(environmentVariable string) (bool, error) {
	if val, ok := os.LookupEnv(environmentVariable); ok {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf(
				"invalid value '%s' for key '%s': expected one of [1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False]",
				val,
				environmentVariable,
			)
		}
		return boolVal, nil
	}
	return false, nil
}
