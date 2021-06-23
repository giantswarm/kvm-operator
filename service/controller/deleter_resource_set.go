package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/wrapper/retryresource"

	"github.com/giantswarm/kvm-operator/service/controller/resource/node"
)

func newDeleterResources(config DeleterConfig) ([]resource.Interface, error) {
	var err error

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
