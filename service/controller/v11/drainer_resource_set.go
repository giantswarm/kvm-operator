package v11

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v11/key"
	"github.com/giantswarm/kvm-operator/service/controller/v11/resource/endpoint"
	"github.com/giantswarm/kvm-operator/service/controller/v11/resource/pod"
)

type DrainerResourceSetConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ProjectName string
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	handlesFunc := func(obj interface{}) bool {
		p, err := key.ToPod(obj)
		if err != nil {
			return false
		}
		v, err := key.VersionBundleVersionFromPod(p)
		// NOTE this is the hack we have to backport to ensure existing clusters
		// work the way they do while paving the ground for v12 where we have to
		// check more explictly against our desired version bundle version.
		if v == "" {
			return true
		}

		// TODO this error handling has to be enabled when going from v11 to v12.
		//if err != nil {
		//	return false
		//}

		// TODO this check has to be enabled when going from v11 to v12.
		//if v == VersionBundle().Version {
		//	return true
		//}

		return false
	}

	var podResource controller.Resource
	{
		c := pod.Config{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		podResource, err = pod.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var endpointResource controller.Resource
	{
		c := endpoint.Config{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := endpoint.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		endpointResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		endpointResource,
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

	var drainerResourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		drainerResourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return drainerResourceSet, nil
}
