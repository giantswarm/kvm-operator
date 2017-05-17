package namespace

import (
	"github.com/juju/errgo"
)

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

var wrongTypeError = errgo.New("wrong type")

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return errgo.Cause(err) == wrongTypeError
}
