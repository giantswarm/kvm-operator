package key

import "github.com/giantswarm/microerror"

var missingAnnotationError = &microerror.Error{
	Kind: "missingAnnotationError",
}

func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
