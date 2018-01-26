package pod

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/framework/podv2"
	"github.com/giantswarm/kvm-operator/service/keyv2"
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

	var resourcesV2 []framework.Resource
	{
		c := podv2.ResourcesConfig{
			Logger:    config.Logger,
			K8sClient: config.K8sClient,

			Name: config.Name,
		}

		resourcesV2, err = podv2.NewResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = config.K8sClient.CoreV1().Pods("")

		c.ListOptions = apismetav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", keyv2.PodWatcherLabel, config.Name),
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var podFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.Informer = newInformer
		c.Logger = config.Logger
		c.ResourceRouter = framework.DefaultResourceRouter(resourcesV2)

		podFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return podFramework, nil
}
