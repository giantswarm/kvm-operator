package clusterrolebinding

import (
	"strings"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/validation"
)

var fieldImmutableError = &microerror.Error{
	Kind: "fieldImmutableError",
}

func isExternalFieldImmutableError(err error) bool {
	return strings.Contains(err.Error(), validation.FieldImmutableErrorMsg)
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
