package kvmconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/kvm-operator/service/keyv3"
)

// newResourceRouter determines which resources are enabled based upon the
// version in the version bundle.
func newResourceRouter(resources map[string][]framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		customObject, err := keyv3.ToCustomObject(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		versionBundleVersion := keyv3.VersionBundleVersion(customObject)
		resourceList, ok := resources[versionBundleVersion]
		if ok {
			return resourceList, nil
		}

		return nil, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", versionBundleVersion)
	}
}
