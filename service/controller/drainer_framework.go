package controller

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v11"
	"github.com/giantswarm/kvm-operator/service/controller/v11/key"
)

type DrainerFrameworkConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ProjectName string
}

func NewDrainerFramework(config DrainerFrameworkConfig) (*framework.Framework, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var err error

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Watcher: config.K8sClient.CoreV1().Pods(""),

			ListOptions: apismetav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", key.PodWatcherLabel, config.ProjectName),
			},
			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceRouter, err := newDrainerResourceRouter(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var drainerFramework *framework.Framework
	{
		c := framework.Config{
			Informer:       newInformer,
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,

			Name: config.ProjectName,
		}

		drainerFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return drainerFramework, nil
}

func newDrainerResourceRouter(config DrainerFrameworkConfig) (*framework.ResourceRouter, error) {
	var err error

	var resourceSetV11 *framework.ResourceSet
	{
		c := v11.DrainerResourceSetConfig{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV11, err = v11.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{
			Logger: config.Logger,

			ResourceSets: []*framework.ResourceSet{
				resourceSetV11,
			},
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceRouter, nil
}
