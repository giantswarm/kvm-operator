// +build k8srequired

package setup

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailed",
}

// IsExecutionFailedError asserts executionFailedError.
func IsExecutionFailedError(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var resourceNotDeleted = &microerror.Error{
	Kind: "resourceNotDeleted",
	Desc: "Resource has not been deleted, but its required to be deleted.",
}

// IsResourceNotDeleted asserts resourceNotDeleted.
func IsResourceNotDeleted(err error) bool {
	return microerror.Cause(err) == resourceNotDeleted
}
