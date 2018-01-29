package podv2

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/kvm-operator/service/resource/podv2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"
)

const (
	ResourceRetries uint64 = 3
)

type ResourcesConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Name is the project name.
	Name string
}

func NewResources(config ResourcesConfig) ([]framework.Resource, error) {
	var err error

	var podResource framework.Resource
	{
		c := podv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		podResource, err = podv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		podResource,
	}

	// Wrap resources with retry and metrics.
	{
		retryWrapConfig := retryresource.DefaultWrapConfig()

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()

		metricsWrapConfig.Name = config.Name

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
