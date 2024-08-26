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
	TitleWriter      io.Writer
	ProcessWriter    io.Writer
	SubprocessWriter io.Writer
	ActionWriter     io.Writer
	DetailWriter     io.Writer
	SubdetailWriter  io.Writer
}

// NewLeveledLogger takes a writer and returns a LeveledLogger that writes to the given
// writer.
func NewLeveledLogger(writer io.Writer) LeveledLogger {
	return LeveledLogger{
		TitleWriter:      NewWriter(writer),
		ProcessWriter:    NewWriter(writer, WithIndent(1)),
		SubprocessWriter: NewWriter(writer, WithIndent(2)),
		ActionWriter:     NewWriter(writer, WithIndent(3)),
		DetailWriter:     NewWriter(writer, WithIndent(4)),
		SubdetailWriter:  NewWriter(writer, WithIndent(5)),
	}
}

// Title takes a string and optional formatting, and prints a formatted string
// with zero levels of indentation.
func (l LeveledLogger) Title(format string, v ...interface{}) {
	l.printf(l.TitleWriter, format, v...)
}

// Process takes a string and optional formatting, and prints a formatted string
// with one level of indentation.
func (l LeveledLogger) Process(format string, v ...interface{}) {
	l.printf(l.ProcessWriter, format, v...)
}

// Subprocess takes a string and optional formatting, and prints a formatted string
// with two levels of indentation.
func (l LeveledLogger) Subprocess(format string, v ...interface{}) {
	l.printf(l.SubprocessWriter, format, v...)
}

// Action takes a string and optional formatting, and prints a formatted string
// with three levels of indentation.
func (l LeveledLogger) Action(format string, v ...interface{}) {
	l.printf(l.ActionWriter, format, v...)
}

// Detail takes a string and optional formatting, and prints a formatted string
// with four levels of indentation.
func (l LeveledLogger) Detail(format string, v ...interface{}) {
	l.printf(l.DetailWriter, format, v...)
}

// Subdetail takes a string and optional formatting, and prints a formatted string
// with five levels of indentation.
func (l LeveledLogger) Subdetail(format string, v ...interface{}) {
	l.printf(l.SubdetailWriter, format, v...)
}

// Break inserts a line break in the log output
func (l LeveledLogger) Break() {
	l.printf(l.TitleWriter, "\n")
}

func (l LeveledLogger) printf(writer io.Writer, format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Fprintf(writer, format, v...)
}
