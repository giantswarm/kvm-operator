package pod

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingClusterLabelError = &microerror.Error{
	Kind: "missingClusterLabelError",
}

// IsMissingClusterLabel asserts  missingClusterLabelError.
func IsMissingClusterLabel(err error) bool {
	return microerror.Cause(err) == missingClusterLabelError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
