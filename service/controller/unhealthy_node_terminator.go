package controller

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v5/pkg/controller"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvm-operator/v4/pkg/label"
	"github.com/giantswarm/kvm-operator/v4/pkg/project"
)

type UnhealthyNodeTerminatorConfig struct {
	K8sClient       k8sclient.Interface
	Logger          micrologger.Logger
	WorkloadCluster workloadcluster.Interface

	ProjectName             string
	TerminateUnhealthyNodes bool
}

type UnhealthyNodeTerminator struct {
	*controller.Controller
}

func NewUnhealthyNodeTerminator(config UnhealthyNodeTerminatorConfig) (*UnhealthyNodeTerminator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	resources, err := newUnhealthyNodeTerminatorResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Resources: resources,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.KVMConfig)
			},
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),

			Name: config.ProjectName + "-unhealthy-nodes-terminator",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	t := &UnhealthyNodeTerminator{
		Controller: operatorkitController,
	}

	return t, nil
}
