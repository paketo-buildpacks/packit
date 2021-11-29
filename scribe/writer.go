package scribe

import (
	"bytes"
	"io"
)

// An Option is a way to configure a writer's format.
type Option func(Writer) Writer

// WithColor takes a Color and returns an Option which can be passed in while
// creating a new Writer to configure the color of the output of the Writer.
func WithColor(color Color) Option {
	return func(l Writer) Writer {
		l.color = color
		return l
	}
}

// WithIndent takes an indent level and returns an Option which can be passed in
// while creating a new Writer to configure the indentation level of the output
// of the Writer.
func WithIndent(indent int) Option {
	return func(l Writer) Writer {
		l.indent = indent
		return l
	}
}

// A Writer conforms to the io.Writer interface and allows for configuration of
// output from the writter such as the color or indentation through Options.
type Writer struct {
	writer io.Writer
	color  Color
	indent int
}

// NewWriter takes a Writer and Options and returns a Writer that will format
// output according to the options given.
func NewWriter(writer io.Writer, options ...Option) Writer {
	w := Writer{writer: writer}
	for _, option := range options {
		w = option(w)
	}

	return w
}

// Write takes the given byte array and formats it in accordance with the
// options on the writer and then outputs that formated text.
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
