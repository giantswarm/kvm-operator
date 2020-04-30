package controller

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

type DrainerConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	CRDLabelSelector string
	ProjectName      string
}

type Drainer struct {
	*controller.Controller
}

func NewDrainer(config DrainerConfig) (*Drainer, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resourceSets, err := newDrainerResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(corev1.Pod)
			},
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			Selector: labels.SelectorFromSet(map[string]string{
				key.PodWatcherLabel: project.Name(),
			}),

			Name: config.ProjectName + "-drainer",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &Drainer{
		Controller: operatorkitController,
	}

	return d, nil
}

func newDrainerResourceSets(config DrainerConfig) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSet, err = NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
