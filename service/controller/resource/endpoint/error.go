package endpoint

import "github.com/giantswarm/microerror"

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

// IsMissingAnnotation asserts missingAnnotationError.
func IsMissingAnnotation(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

var missingClusterLabelError = &microerror.Error{
	Kind: "missingClusterLabelError",
}

// IsMissingClusterLabel asserts  missingClusterLabelError.
func IsMissingClusterLabel(err error) bool {
	return microerror.Cause(err) == missingClusterLabelError
}
