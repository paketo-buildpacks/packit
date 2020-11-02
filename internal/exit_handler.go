package internal

import (
	"fmt"
	"io"
	"os"
)

type Option func(handler ExitHandler) ExitHandler

func WithExitHandlerStderr(stderr io.Writer) Option {
	return func(handler ExitHandler) ExitHandler {
		handler.stderr = stderr
		return handler
	}
}

func WithExitHandlerStdout(stdout io.Writer) Option {
	return func(handler ExitHandler) ExitHandler {
		handler.stdout = stdout
		return handler
	}
}

func WithExitHandlerExitFunc(e func(int)) Option {
	return func(handler ExitHandler) ExitHandler {
		handler.exitFunc = e
		return handler
	}
}

type ExitHandler struct {
	stdout   io.Writer
	stderr   io.Writer
	exitFunc func(int)
}

func NewExitHandler(options ...Option) ExitHandler {
	handler := ExitHandler{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		exitFunc: os.Exit,
	}

	for _, option := range options {
		handler = option(handler)
	}

	return handler
}

func (h ExitHandler) Error(err error) {
	fmt.Fprintln(h.stderr, err)

	var code int
	switch err.(type) {
	case failError:
		code = 100
	case nil:
		code = 0
	default:
		code = 1
	}

	h.exitFunc(code)
}
