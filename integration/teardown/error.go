package teardown

import "github.com/giantswarm/microerror"

var resourceNotDeleted = &microerror.Error{
	Kind: "resourceNotDeleted",
}

// IsInvalidConfig asserts invalidConfigError.
func IsResourceNotDeleted(err error) bool {
	return microerror.Cause(err) == resourceNotDeleted
}
