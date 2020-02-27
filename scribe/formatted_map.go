package scribe

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudfoundry/packit"
)

type FormattedMap map[string]interface{}

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

func NewFormattedMapFromEnvironment(environment packit.Environment) FormattedMap {
	envMap := FormattedMap{}
	for key, value := range environment {
		parts := strings.SplitN(key, ".", 2)

		switch {
		case parts[1] == "override" || parts[1] == "default":
			envMap[parts[0]] = value
		case parts[1] == "prepend":
			delim := environment[parts[0]+".delim"]
			envMap[parts[0]] = fmt.Sprintf("%s", strings.Join([]string{value, "$" + parts[0]}, delim))
		case parts[1] == "append":
			delim := environment[parts[0]+".delim"]
			envMap[parts[0]] = fmt.Sprintf("%s", strings.Join([]string{"$" + parts[0], value}, delim))
		}
	}

	return envMap
}
