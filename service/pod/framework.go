package pod

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/pod/v2"
	"github.com/giantswarm/kvm-operator/service/pod/v2/key"
)

type FrameworkConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Name is the name of the project.
	Name string
}

func NewFramework(config FrameworkConfig) (*framework.Framework, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}

	v2ResourceSetHandles := func(obj interface{}) bool {
		return true
	}

	var v2Resources []framework.Resource
	{
		c := v2.ResourcesConfig{
			Logger:    config.Logger,
			K8sClient: config.K8sClient,

			Name: config.Name,
		}

		v2Resources, err = v2.NewResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v2ResourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{}

		c.Handles = v2ResourceSetHandles
		c.Logger = config.Logger
		c.Resources = v2Resources

		v2ResourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Watcher: config.K8sClient.CoreV1().Pods(""),

			ListOptions: apismetav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", key.PodWatcherLabel, config.Name),
			},
			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{
			Logger: config.Logger,

			ResourceSets: []*framework.ResourceSet{
				v2ResourceSet,
			},
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var podFramework *framework.Framework
	{
		c := framework.Config{}

		c.Informer = newInformer
		c.Logger = config.Logger
		c.ResourceRouter = resourceRouter

		podFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return podFramework, nil
}
