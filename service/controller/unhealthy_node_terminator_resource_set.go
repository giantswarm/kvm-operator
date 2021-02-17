package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"

	"github.com/giantswarm/kvm-operator/service/controller/resource/terminateunhealthynodes"
)

func newUnhealthyNodeTerminatorResources(config UnhealthyNodeTerminatorConfig) ([]resource.Interface, error) {
	var err error
	var terminateUnhealthyNodesResource resource.Interface
	{
		c := terminateunhealthynodes.Config{
			K8sClient:       config.K8sClient.K8sClient(),
			Logger:          config.Logger,
			WorkloadCluster: config.WorkloadCluster,
		}

		terminateUnhealthyNodesResource, err = terminateunhealthynodes.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		terminateUnhealthyNodesResource,
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
