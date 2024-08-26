package scribe

import (
	"fmt"
	"sort"
	"strings"
)

// A FormattedMap is a wrapper for map[string]interface{} to extend functionality.
type FormattedMap map[string]interface{}

// Sorts all of the keys in the FormattedMap alphabetically and then constructs
// a padded table.
func (m FormattedMap) String() string {
	var (
		keys    []string
		padding int
	)
	for key := range m {
		if len(key) > padding {
			padding = len(key)
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		value := m[key]
		if value == nil {
			value = "<empty>"
		}

		for i := len(key); i < padding; i++ {
			key = key + " "
		}

		builder.WriteString(fmt.Sprintf("%s -> \"%v\"\n", key, value))
	}

	return strings.TrimSpace(builder.String())
}

// NewFormattedMapFromEnvironment take an environment and returns a
// FormattedMap with the appropriate environment variable information added.
func NewFormattedMapFromEnvironment(environment map[string]string) FormattedMap {
	envMap := FormattedMap{}
	for key, value := range environment {
		parts := strings.SplitN(key, ".", 2)

		if len(parts) < 2 {
			envMap[key] = value
			continue
		}

		switch {
		case parts[1] == "override" || parts[1] == "default":
			envMap[parts[0]] = value
		case parts[1] == "prepend":
			existingValue, ok := envMap[parts[0]]
			if !ok {
				existingValue = "$" + parts[0]
			}
			envMap[parts[0]] = strings.Join([]string{value, fmt.Sprintf("%v", existingValue)}, environment[parts[0]+".delim"])
		case parts[1] == "append":
			existingValue, ok := envMap[parts[0]]
			if !ok {
				existingValue = "$" + parts[0]
			}
			envMap[parts[0]] = strings.Join([]string{fmt.Sprintf("%v", existingValue), value}, environment[parts[0]+".delim"])
		}
	}

	return envMap
}
