package nodecontroller

import (
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller/resource/nodeready"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

type Config struct {
	Cluster             v1alpha1.KVMConfig
	ManagementK8sClient client.Client
	WorkloadK8sClient   k8sclient.Interface
	Logger              micrologger.Logger
}

type Controller struct {
	*controller.Controller
}

func New(config Config) (*Controller, error) {
	var err error

	resources, err := newResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.WorkloadK8sClient,
			Logger:    config.Logger,
			Resources: resources,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(corev1.Node)
			},
			Selector: labels.Everything(),

			Name: fmt.Sprintf("%s-%s-nodes", project.Name(), key.ClusterID(config.Cluster)),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Controller{
		Controller: operatorkitController,
	}

	return c, nil
}

func newResources(config Config) ([]resource.Interface, error) {
	var readyLabelResource resource.Interface
	{
		c := nodeready.Config{
			Cluster:             config.Cluster,
			ManagementK8sClient: config.ManagementK8sClient,
			Logger:              config.Logger,
		}

		var err error
		readyLabelResource, err = nodeready.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		readyLabelResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		var err error
		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
