package deploymentv3

import "github.com/giantswarm/microerror"

var emptyAnnotationError = microerror.New("empty annotation")

// IsEmptyAnnotation asserts emptyAnnotationError.
func IsEmptyAnnotation(err error) bool {
	return microerror.Cause(err) == emptyAnnotationError
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

var missingAnnotationError = microerror.New("missing annotation")

// IsMissingAnnotation asserts missingAnnotationError.
func IsMissingAnnotation(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
