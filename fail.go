package packit

import "github.com/paketo-buildpacks/packit/internal"

// Fail is a sentinal value that can be used to indicate a failure to detect
// during the detect phase. Fail implements the Error interface and should be
// returned as the error value in the DetectFunc signature. Fail also supports
// a modifier function, WithMessage, that allows the caller to set a custom
// failure message. The WithMessage function supports a fmt.Printf-like format
// string and variadic arguments to build a message, eg:
// packit.Fail.WithMessage("failed: %w", err).
var Fail = internal.Fail
