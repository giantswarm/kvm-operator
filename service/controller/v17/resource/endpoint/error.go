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

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
