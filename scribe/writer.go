package scribe

import (
	"bytes"
	"io"
)

type Option func(Writer) Writer

func WithColor(color Color) Option {
	return func(l Writer) Writer {
		l.color = color
		return l
	}
}

func WithIndent(indent int) Option {
	return func(l Writer) Writer {
		l.indent = indent
		return l
	}
}

type Writer struct {
	writer io.Writer
	color  Color
	indent int
}

func NewWriter(writer io.Writer, options ...Option) Writer {
	w := Writer{writer: writer}
	for _, option := range options {
		w = option(w)
	}

	return w
}

func (w Writer) Write(b []byte) (int, error) {
	var (
		prefix, suffix []byte
		reset          = []byte("\r")
		newline        = []byte("\n")
		n              = len(b)
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
		for i := 0; i < w.indent; i++ {
			line = append([]byte("  "), line...)
		}
		indentedLines = append(indentedLines, line)
	}

	b = bytes.Join(indentedLines, newline)

	if w.color != nil {
		b = []byte(w.color(string(b)))
	}

	if prefix != nil {
		b = append(prefix, b...)
	}

	if suffix != nil {
		b = append(b, suffix...)
	}

	_, err := w.writer.Write(b)
	if err != nil {
		return n, err
	}

	return n, nil
}
