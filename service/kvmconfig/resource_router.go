package kvmconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/key"
)

// newResourceRouter determines which resources are enabled based upon the
// version in the version bundle.
func newResourceRouter(resources map[string][]framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		versionBundleVersion := key.VersionBundleVersion(customObject)
		resourceList, ok := resources[versionBundleVersion]
		if ok {
			return resourceList, nil
		}

		return nil, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", versionBundleVersion)
	}
}
