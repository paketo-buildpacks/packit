package scribe

import (
	"fmt"
	"sort"
	"strings"
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

		builder.WriteString(fmt.Sprintf("%s -> %v\n", key, value))
	}

	return strings.TrimSpace(builder.String())
}
