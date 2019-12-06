package controller

import (
	"fmt"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvm-operator/pkg/project"
	v20 "github.com/giantswarm/kvm-operator/service/controller/v20"
	"github.com/giantswarm/kvm-operator/service/controller/v20/key"
	v21 "github.com/giantswarm/kvm-operator/service/controller/v21"
	v22 "github.com/giantswarm/kvm-operator/service/controller/v22"
	v23 "github.com/giantswarm/kvm-operator/service/controller/v23"
	"github.com/giantswarm/kvm-operator/service/controller/v23patch1"
	v24 "github.com/giantswarm/kvm-operator/service/controller/v24"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1"
	v25 "github.com/giantswarm/kvm-operator/service/controller/v25"
	v26 "github.com/giantswarm/kvm-operator/service/controller/v26"
)

type DrainerConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	CRDLabelSelector string
	ProjectName      string
}

func (c DrainerConfig) newInformerListOptions() metav1.ListOptions {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", key.PodWatcherLabel, c.ProjectName),
	}

	if c.CRDLabelSelector != "" {
		listOptions.LabelSelector = listOptions.LabelSelector + "," + c.CRDLabelSelector
	}

	return listOptions
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
			Logger:    config.Logger,
			MatchLabels: map[string]string{
				key.PodWatcherLabel: project.Name(),
			},
			ResourceSets: resourceSets,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(corev1.Pod)
			},

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

	var resourceSetV20 *controller.ResourceSet
	{
		c := v20.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV20, err = v20.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV21 *controller.ResourceSet
	{
		c := v21.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV21, err = v21.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV22 *controller.ResourceSet
	{
		c := v22.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV22, err = v22.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23 *controller.ResourceSet
	{
		c := v23.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV23, err = v23.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23patch1 *controller.ResourceSet
	{
		c := v23patch1.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV23patch1, err = v23patch1.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV24 *controller.ResourceSet
	{
		c := v24.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV24, err = v24.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV24patch1 *controller.ResourceSet
	{
		c := v24patch1.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

<<<<<<< HEAD
		resourceSetV24patch1, err = v24patch1.NewDrainerResourceSet(c)
=======
		resourceSetV24, err = v24patch1.NewDrainerResourceSet(c)
>>>>>>> d6f149c2... wire v24patch1
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV25 *controller.ResourceSet
	{
		c := v25.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV25, err = v25.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV26 *controller.ResourceSet
	{
		c := v26.DrainerResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV26, err = v26.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSetV20,
		resourceSetV21,
		resourceSetV22,
		resourceSetV23,
		resourceSetV23patch1,
		resourceSetV24,
		resourceSetV24patch1,
		resourceSetV25,
		resourceSetV26,
	}

	return resourceSets, nil
}
