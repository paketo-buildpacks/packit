package scribe

import (
	"fmt"
	"io"
	"strings"
)

type Logger struct {
	writer io.Writer
	LeveledLogger
	Debug LeveledLogger
}

func NewLogger(writer io.Writer) Logger {
	return Logger{
		writer:        writer,
		LeveledLogger: NewLeveledLogger(writer),
		Debug:         NewLeveledLogger(io.Discard),
	}
}

func (l Logger) WithLevel(level string) Logger {
	switch level {
	case "DEBUG":
		return Logger{
			writer:        l.writer,
			LeveledLogger: NewLeveledLogger(l.writer),
			Debug:         NewLeveledLogger(l.writer),
		}
	default:
		return Logger{
			writer:        l.writer,
			LeveledLogger: NewLeveledLogger(l.writer),
			Debug:         NewLeveledLogger(io.Discard),
		}
	}
}

type LeveledLogger struct {
	title      io.Writer
	process    io.Writer
	subprocess io.Writer
	action     io.Writer
	detail     io.Writer
	subdetail  io.Writer
}

func NewLeveledLogger(writer io.Writer) LeveledLogger {
	return LeveledLogger{
		title:      NewWriter(writer),
		process:    NewWriter(writer, WithIndent(1)),
		subprocess: NewWriter(writer, WithIndent(2)),
		action:     NewWriter(writer, WithIndent(3)),
		detail:     NewWriter(writer, WithIndent(4)),
		subdetail:  NewWriter(writer, WithIndent(5)),
	}
}

func (l LeveledLogger) Title(format string, v ...interface{}) {
	l.printf(l.title, format, v...)
}

func (l LeveledLogger) Process(format string, v ...interface{}) {
	l.printf(l.process, format, v...)
}

func (l LeveledLogger) Subprocess(format string, v ...interface{}) {
	l.printf(l.subprocess, format, v...)
}

func (l LeveledLogger) Action(format string, v ...interface{}) {
	l.printf(l.action, format, v...)
}

func (l LeveledLogger) Detail(format string, v ...interface{}) {
	l.printf(l.detail, format, v...)
}

func (l LeveledLogger) Subdetail(format string, v ...interface{}) {
	l.printf(l.subdetail, format, v...)
}

func (l LeveledLogger) Break() {
	l.printf(l.title, "\n")
}

func (l LeveledLogger) printf(writer io.Writer, format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Fprintf(writer, format, v...)
}
