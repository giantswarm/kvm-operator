package service

import (
	"context"

	"github.com/giantswarm/kvm-operator/service/keyv2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// NewResourceRouter determines which resources are enabled based upon the
// version in the version bundle.
func NewResourceRouter(resources map[string][]framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		var enabledResources []framework.Resource

		customObject, err := keyv2.ToCustomObject(obj)
		if err != nil {
			return enabledResources, microerror.Mask(err)
		}

		switch keyv2.VersionBundleVersion(customObject) {
		case keyv2.K8sCloudConfig_V_1_1_0:
			// Legacy version so only enable the legacy resource.
			enabledResources = resources[keyv2.K8sCloudConfig_V_1_1_0]
		case keyv2.K8sCloudConfig_V_2_0_0:
			// Cloud Formation transitional version so enable all resources.
			enabledResources = resources[keyv2.K8sCloudConfig_V_2_0_0]
		case "":
			// Default to the legacy resource for custom objects without a version
			// bundle.
			enabledResources = resources[keyv2.K8sCloudConfig_V_1_1_0]
		default:
			return enabledResources, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", keyv2.VersionBundleVersion(customObject))
		}

		return enabledResources, nil
	}
}
