package controller

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
)

type TransitionConfig struct {
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

type Transition struct {
	*controller.Controller
}

func NewTransition(config TransitionConfig) (*Transition, error) {
	resources, err := newTransitionResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.KVMConfig)
			},
			Logger:    config.Logger,
			Resources: resources,
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),

			Name: project.Name() + "-transition",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &Transition{
		Controller: operatorkitController,
	}

	return d, nil
}
