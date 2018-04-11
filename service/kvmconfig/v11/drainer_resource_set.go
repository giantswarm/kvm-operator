package v11

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v11/resource/pod"
)

type DrainerResourceSetConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ProjectName string
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*framework.ResourceSet, error) {
	var err error

	handlesFunc := func(obj interface{}) bool {
		return true
	}

	var podResource framework.Resource
	{
		c := pod.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		podResource, err = pod.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		podResource,
	}

	{
		c := retryresource.WrapConfig{
			BackOffFactory: func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), uint64(3)) },
			Logger:         config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerResourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{
			Handles:   handlesFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		drainerResourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return drainerResourceSet, nil
}
