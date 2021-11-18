package scribe

import (
	"sort"
	"strings"
)

// A FormattedList is a wrapper for []string to extend functionality.
type FormattedList []string

// Sorts the FormattedList alphabetically and then prints each item on its own
// line.
func (l FormattedList) String() string {
	sort.Strings(l)

	var builder strings.Builder
	for i, elem := range l {
		builder.WriteString(elem)
		if i < len(l)-1 {
			builder.WriteRune('\n')
		}
	}

	return builder.String()
}
