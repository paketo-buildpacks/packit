package scribe

import (
	"sort"
	"strings"
)

type FormattedList []string

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
