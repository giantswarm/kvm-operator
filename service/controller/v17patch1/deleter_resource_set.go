package v17patch1

import (
	"github.com/giantswarm/certs"
	"github.com/giantswarm/kvm-operator/service/controller/v17patch1/key"
	"github.com/giantswarm/kvm-operator/service/controller/v17patch1/resource/node"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"
)

type DeleterResourceSetConfig struct {
	CertsSearcher certs.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger

	ProjectName string
}

func NewDeleterResourceSet(config DeleterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	handlesFunc := func(obj interface{}) bool {
		kvmConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(kvmConfig) == VersionBundle().Version {
			return true
		}

		return false
	}

	var nodeResource controller.Resource
	{
		c := node.Config{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
		}

		nodeResource, err = node.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
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

	var deleterResourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		deleterResourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return deleterResourceSet, nil
}
