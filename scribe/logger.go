package scribe

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Logger struct {
	title      io.Writer
	process    io.Writer
	subprocess io.Writer
	action     io.Writer
	detail     io.Writer
	subdetail  io.Writer

	debug bool
	info  bool
}

type infoLogger struct {
	logger Logger

	debug bool
	info  bool
}

type debugLogger struct {
	logger Logger

	debug bool
}

func NewLogger(writer io.Writer) Logger {
	l := Logger{
		title:      NewWriter(writer),
		process:    NewWriter(writer, WithIndent(1)),
		subprocess: NewWriter(writer, WithIndent(2)),
		action:     NewWriter(writer, WithIndent(3)),
		detail:     NewWriter(writer, WithIndent(4)),
		subdetail:  NewWriter(writer, WithIndent(5)),
	}

	if level, ok := os.LookupEnv("BP_LOG_LEVEL"); ok {
		switch level {
		case "INFO":
			l.info = true
		case "DEBUG":
			l.debug = true
		}
	}

	return l
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

func (l Logger) Info() infoLogger {
	return infoLogger{
		logger: l,

		debug: l.debug,
		info:  l.info,
	}
}

func (il infoLogger) Title(format string, v ...interface{}) {
	if il.info || il.debug {
		il.logger.Title(format, v...)
	}
}

func (il infoLogger) Process(format string, v ...interface{}) {
	if il.info || il.debug {
		il.logger.Process(format, v...)
	}
}

func (il infoLogger) Subprocess(format string, v ...interface{}) {
	if il.info || il.debug {
		il.logger.Subprocess(format, v...)
	}
}

func (il infoLogger) Action(format string, v ...interface{}) {
	if il.info || il.debug {
		il.logger.Action(format, v...)
	}
}

func (il infoLogger) Detail(format string, v ...interface{}) {
	if il.info || il.debug {
		il.logger.Detail(format, v...)
	}
}

func (il infoLogger) Subdetail(format string, v ...interface{}) {
	if il.info || il.debug {
		il.logger.Subdetail(format, v...)
	}
}

func (il infoLogger) Break() {
	if il.info || il.debug {
		il.logger.Break()
	}
}

func (l Logger) Debug() debugLogger {
	return debugLogger{
		logger: l,

		debug: l.debug,
	}
}

func (dl debugLogger) Title(format string, v ...interface{}) {
	if dl.debug {
		dl.logger.Title(format, v...)
	}
}

func (dl debugLogger) Process(format string, v ...interface{}) {
	if dl.debug {
		dl.logger.Process(format, v...)
	}
}

func (dl debugLogger) Subprocess(format string, v ...interface{}) {
	if dl.debug {
		dl.logger.Subprocess(format, v...)
	}
}

func (dl debugLogger) Action(format string, v ...interface{}) {
	if dl.debug {
		dl.logger.Action(format, v...)
	}
}

func (dl debugLogger) Detail(format string, v ...interface{}) {
	if dl.debug {
		dl.logger.Detail(format, v...)
	}
}

func (dl debugLogger) Subdetail(format string, v ...interface{}) {
	if dl.debug {
		dl.logger.Subdetail(format, v...)
	}
}

func (dl debugLogger) Break() {
	if dl.debug {
		dl.logger.Break()
	}
}

func (l Logger) printf(writer io.Writer, format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Fprintf(writer, format, v...)
}
