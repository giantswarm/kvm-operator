// +build k8srequired

package teardown

import "github.com/giantswarm/microerror"

var resourceNotDeleted = &microerror.Error{
	Kind: "resourceNotDeleted",
	Desc: "Resource has not been deleted, but its required to be deleted.",
}

// IsResourceNotDeleted asserts resourceNotDeleted.
func IsResourceNotDeleted(err error) bool {
	return microerror.Cause(err) == resourceNotDeleted
}
