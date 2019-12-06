<<<<<<< HEAD
package v24patch1
=======
package v24
>>>>>>> c4c6c79d... copy v24 to v24patch1

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"

<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/node"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/node"
>>>>>>> c4c6c79d... copy v24 to v24patch1
)

type DeleterResourceSetConfig struct {
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface

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

	var nodeResource resource.Interface
	{
		c := node.Config{
			K8sClient:     config.K8sClient,
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
