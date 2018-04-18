package pod

import "github.com/giantswarm/microerror"

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingAnnotationError = microerror.New("missing annotation")

func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
