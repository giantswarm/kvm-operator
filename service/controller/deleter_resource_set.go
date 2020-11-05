package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"

	"github.com/giantswarm/kvm-operator/service/controller/resource/cleanupendpointips"
	"github.com/giantswarm/kvm-operator/service/controller/resource/node"
)

func newDeleterResources(config DeleterConfig) ([]resource.Interface, error) {
	var err error

	var cleanupendpointipsResource resource.Interface
	{
		c := cleanupendpointips.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,
		}

		cleanupendpointipsResource, err = cleanupendpointips.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var nodeResource resource.Interface
	{
		c := node.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,
		}

		nodeResource, err = node.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		cleanupendpointipsResource,
		nodeResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
