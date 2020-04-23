package clusterrolebinding

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var fieldImmutableError = &microerror.Error{
	Kind: "fieldImmutableError",
}

func isExternalFieldImmutableError(err error) bool {
	// Would be nice to use validation.FieldImmutableErrorMsg here,
	// but can't seem to get it from the wrapped error reliably
	return strings.Contains(err.Error(), "cannot change roleRef")
}

// IsFieldImmutableError asserts fieldImmutableError
func IsFieldImmutableError(err error) bool {
	return microerror.Cause(err) == fieldImmutableError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
