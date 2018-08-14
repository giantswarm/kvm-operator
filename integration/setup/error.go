package setup

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailed",
}

// IsExecutionFailedError asserts executionFailedError.
func IsExecutionFailedError(err error) bool {
	return microerror.Cause(err) == executionFailedError
}
