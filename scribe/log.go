package scribe

import (
	"bytes"
	"fmt"
	"io"
)

type Option func(Log) Log

func WithColor(color Color) Option {
	return func(l Log) Log {
		l.color = color
		return l
	}
}

func WithIndent(indent int) Option {
	return func(l Log) Log {
		l.indent = indent
		return l
	}
}

type Log struct {
	w      io.Writer
	color  Color
	indent int
}

func NewLog(w io.Writer, options ...Option) Log {
	log := Log{w: w}
	for _, option := range options {
		log = option(log)
	}

	return log
}

func (l Log) Write(b []byte) (int, error) {
	var (
		prefix, suffix []byte
		reset          = []byte("\r")
		newline        = []byte("\n")
	)

	if bytes.HasPrefix(b, reset) {
		b = bytes.TrimPrefix(b, reset)
		prefix = reset
	}

	if bytes.HasSuffix(b, newline) {
		b = bytes.TrimSuffix(b, newline)
		suffix = newline
	}

	lines := bytes.Split(b, newline)

	var indentedLines [][]byte
	for _, line := range lines {
		for i := 0; i < l.indent; i++ {
			line = append([]byte("  "), line...)
		}
		indentedLines = append(indentedLines, line)
	}

	b = bytes.Join(indentedLines, newline)

	if l.color != nil {
		b = []byte(l.color(string(b)))
	}

	if prefix != nil {
		b = append(prefix, b...)
	}

	if suffix != nil {
		b = append(b, suffix...)
	}

	return l.w.Write(b)
}

func (l Log) Print(v ...interface{}) {
	fmt.Fprint(l, v...)
}

func (l Log) Println(v ...interface{}) {
	fmt.Fprintln(l, v...)
}

func (l Log) Printf(format string, v ...interface{}) {
	fmt.Fprintf(l, format, v...)
}

func (l Log) Break() {
	l.Println()
}
