package cleanupendpointips

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

var noPodForNodeError = &microerror.Error{
	Kind: "noPodForNodeError",
}

// IsNoPodForNodeError asserts noPodForNodeError
func IsNoPodForNodeError(err error) bool {
	return microerror.Cause(err) == noPodForNodeError
}
