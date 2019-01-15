package endpoint

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingAnnotationError = &microerror.Error{
	Kind: "missingAnnotationError",
}

func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var serviceNotFoundError = &microerror.Error{
	Kind: "serviceNotFoundError",
}

// IsServiceNotFound asserts serviceNotFoundError.
func IsServiceNotFound(err error) bool {
	return microerror.Cause(err) == serviceNotFoundError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
