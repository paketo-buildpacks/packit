package scribe

import (
	"strconv"
)

var (
	BlackColor   = NewColor(false, 0, -1)
	RedColor     = NewColor(false, 1, -1)
	GreenColor   = NewColor(false, 2, -1)
	YellowColor  = NewColor(false, 3, -1)
	BlueColor    = NewColor(false, 4, -1)
	MagentaColor = NewColor(false, 5, -1)
	CyanColor    = NewColor(false, 6, -1)
	WhiteColor   = NewColor(false, 7, -1)
	GrayColor    = NewColor(false, 244, -1)
)

type Color func(message string) string

func NewColor(bold bool, fg, bg int) Color {
	return func(message string) string {
		prefix := "\x1b["
		if bold {
			prefix = prefix + "1"
		} else {
			prefix = prefix + "0"
		}

		if fg >= 0 {
			prefix = prefix + ";38;5;" + strconv.Itoa(fg)
		}

		if bg >= 0 {
			prefix = prefix + ";48;5;" + strconv.Itoa(bg)
		}

		prefix = prefix + "m"
		suffix := "\x1b[0m"

		return prefix + message + suffix
	}
}
