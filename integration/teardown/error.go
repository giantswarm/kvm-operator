package teardown

import "github.com/giantswarm/microerror"

var resourceNotDeleted = &microerror.Error{
	Kind: "resourceNotDeleted",
	Desc: "Resource has not been deleted, but its required to be deleted.",
}

// IsInvalidConfig asserts invalidConfigError.
func IsResourceNotDeleted(err error) bool {
	return microerror.Cause(err) == resourceNotDeleted
}
