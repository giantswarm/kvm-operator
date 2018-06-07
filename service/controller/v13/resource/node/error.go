package node

import (
	"github.com/giantswarm/microerror"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

var instanceNotFoundError = microerror.New("instance not found")

// IsInstanceNotFound asserts instanceNotFoundError.
func IsInstanceNotFound(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if c == instanceNotFoundError {
		return true
	}

	if c == cloudprovider.InstanceNotFound {
		return true
	}

	return false
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
