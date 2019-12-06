<<<<<<< HEAD
package v24patch1
=======
package v24
>>>>>>> c4c6c79d... copy v24 to v24patch1

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
