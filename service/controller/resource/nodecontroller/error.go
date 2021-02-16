package nodecontroller

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

// A regular expression representing arbitrary internal errors from a Kubernetes API server.
var serverErrorPattern = regexp.MustCompile(`an error on the server \(.*\) has prevented the request from succeeding`)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

// isServerError return true if the error represents an arbitrary internal error from a Kubernetes API server.
func isServerError(err error) bool {
	return serverErrorPattern.MatchString(err.Error())
}
