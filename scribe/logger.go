package scribe

import (
	"fmt"
	"io"
	"strings"
)

type Logger struct {
	title      io.Writer
	process    io.Writer
	subprocess io.Writer
	action     io.Writer
	detail     io.Writer
	subdetail  io.Writer
}

func NewLogger(writer io.Writer) Logger {
	return Logger{
		title:      NewWriter(writer),
		process:    NewWriter(writer, WithIndent(1)),
		subprocess: NewWriter(writer, WithIndent(2)),
		action:     NewWriter(writer, WithIndent(3)),
		detail:     NewWriter(writer, WithIndent(4)),
		subdetail:  NewWriter(writer, WithIndent(5)),
	}
}

func (l Logger) Title(format string, v ...interface{}) {
	l.printf(l.title, format, v...)
}

func (l Logger) Process(format string, v ...interface{}) {
	l.printf(l.process, format, v...)
}

func (l Logger) Subprocess(format string, v ...interface{}) {
	l.printf(l.subprocess, format, v...)
}

func (l Logger) Action(format string, v ...interface{}) {
	l.printf(l.action, format, v...)
}

func (l Logger) Detail(format string, v ...interface{}) {
	l.printf(l.detail, format, v...)
}

func (l Logger) Subdetail(format string, v ...interface{}) {
	l.printf(l.subdetail, format, v...)
}

func (l Logger) Break() {
	l.printf(l.title, "\n")
}

func (l Logger) printf(writer io.Writer, format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Fprintf(writer, format, v...)
}
