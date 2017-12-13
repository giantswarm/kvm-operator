package service

import (
	"context"

	"github.com/giantswarm/kvm-operator/service/keyv2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// newResourceRouter determines which resources are enabled based upon the
// version in the version bundle.
func newResourceRouter(resources []framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		var enabledResources []framework.Resource

		customObject, err := keyv2.ToCustomObject(obj)
		if err != nil {
			return enabledResources, microerror.Mask(err)
		}

		switch keyv2.VersionBundleVersion(customObject) {
		case "0.1.0":
			enabledResources = resources
		case "1.0.0":
			enabledResources = resources
		default:
			return enabledResources, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", keyv2.VersionBundleVersion(customObject))
		}

		return enabledResources, nil
	}
}
