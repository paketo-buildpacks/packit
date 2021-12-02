package scribe

import (
	"fmt"
	"io"
	"strings"
)

// A Logger provides a standard logging interface for doing basic low level
// logging tasks as well as debug logging.
type Logger struct {
	writer io.Writer
	LeveledLogger
	Debug LeveledLogger
}

// NewLogger takes a writer and returns a Logger that writes to the given
// writer. The default writter sends all debug logging to io.Discard.
func NewLogger(writer io.Writer) Logger {
	return Logger{
		writer:        writer,
		LeveledLogger: NewLeveledLogger(writer),
		Debug:         NewLeveledLogger(io.Discard),
	}
}

// WithLevel takes in a log level string and configures the log level of the
// logger. To enable debug logging the log level must be set to "DEBUG".
func (l Logger) WithLevel(level string) Logger {
	switch strings.ToUpper(level) {
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

// A LeveledLogger provides a standard interface for basic formatted logging.
type LeveledLogger struct {
	title      io.Writer
	process    io.Writer
	subprocess io.Writer
	action     io.Writer
	detail     io.Writer
	subdetail  io.Writer
}

// NewLeveledLogger takes a writer and returns a LeveledLogger that writes to the given
// writer.
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

// Title takes a string and optional formatting, and prints a formatted string
// with zero levels of indentation.
func (l LeveledLogger) Title(format string, v ...interface{}) {
	l.printf(l.title, format, v...)
}

// Process takes a string and optional formatting, and prints a formatted string
// with one level of indentation.
func (l LeveledLogger) Process(format string, v ...interface{}) {
	l.printf(l.process, format, v...)
}

// Subprocess takes a string and optional formatting, and prints a formatted string
// with two levels of indentation.
func (l LeveledLogger) Subprocess(format string, v ...interface{}) {
	l.printf(l.subprocess, format, v...)
}

// Action takes a string and optional formatting, and prints a formatted string
// with three levels of indentation.
func (l LeveledLogger) Action(format string, v ...interface{}) {
	l.printf(l.action, format, v...)
}

// Detail takes a string and optional formatting, and prints a formatted string
// with four levels of indentation.
func (l LeveledLogger) Detail(format string, v ...interface{}) {
	l.printf(l.detail, format, v...)
}

// Subdetail takes a string and optional formatting, and prints a formatted string
// with five levels of indentation.
func (l LeveledLogger) Subdetail(format string, v ...interface{}) {
	l.printf(l.subdetail, format, v...)
}

// Break inserts a line break in the log output
func (l LeveledLogger) Break() {
	l.printf(l.title, "\n")
}

func (l LeveledLogger) printf(writer io.Writer, format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Fprintf(writer, format, v...)
}
