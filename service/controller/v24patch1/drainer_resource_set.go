<<<<<<< HEAD
<<<<<<< HEAD
package v24patch1
=======
package v24
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
package v24patch1
>>>>>>> d6f149c2... wire v24patch1

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"k8s.io/client-go/kubernetes"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/endpoint"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/pod"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/endpoint"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/pod"
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/endpoint"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/pod"
>>>>>>> d6f149c2... wire v24patch1
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
		if err != nil {
			return false
		}

		if v == VersionBundle().Version {
			return true
		}

		return false
	}

	var podResource resource.Interface
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

	var endpointResource resource.Interface
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
