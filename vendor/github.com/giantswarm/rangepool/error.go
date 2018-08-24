package rangepool

import (
	"github.com/giantswarm/microerror"
)

var capacityReachedError = microerror.New("capacity reached")

// IsCapacityReached asserts capacityReachedError.
func IsCapacityReached(err error) bool {
	return microerror.Cause(err) == capacityReachedError
}

var executionFailedError = microerror.New("execution failed")

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var itemsNotFoundError = microerror.New("items not found")

// IsItemsNotFound asserts itemsNotFoundError.
func IsItemsNotFound(err error) bool {
	return microerror.Cause(err) == itemsNotFoundError
}
