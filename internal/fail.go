package internal

import (
	"errors"
	"fmt"
)

var Fail = failError{error: errors.New("failed")}

type failError struct {
	error
}

func (f failError) WithMessage(format string, v ...interface{}) failError {
	return failError{error: fmt.Errorf(format, v...)}
}
