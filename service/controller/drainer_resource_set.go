package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"

	"github.com/giantswarm/kvm-operator/service/controller/resource/endpoint"
	"github.com/giantswarm/kvm-operator/service/controller/resource/pod"
)

func newDrainerResources(config DrainerConfig) ([]resource.Interface, error) {
	var err error

	var endpointResource resource.Interface
	{
		c := endpoint.Config{
			G8sClient:  config.K8sClient.G8sClient(),
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		endpointResource, err = endpoint.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var podResource resource.Interface
	{
		c := pod.Config{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		podResource, err = pod.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		endpointResource,
		podResource,
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
