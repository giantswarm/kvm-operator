package kvmconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v2"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v4"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v5"
)

type FrameworkConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestUpdateEnabled bool
	// Name is the name of the project.
	Name string
}

func NewFramework(config FrameworkConfig) (*framework.Framework, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.K8sExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sExtClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}

	var err error

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = config.K8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certsSearcher certs.Interface
	{
		c := certs.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomkeysSearcher randomkeys.Interface
	{
		keyConfig := randomkeys.DefaultConfig()
		keyConfig.K8sClient = config.K8sClient
		keyConfig.Logger = config.Logger
		randomkeysSearcher, err = randomkeys.NewSearcher(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = config.G8sClient.ProviderV1alpha1().KVMConfigs("")

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV2 *framework.ResourceSet
	{
		c := v2.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			HandledVersionBundles: []string{
				"1.0.0",
				"0.1.0",
				"", // This is for legacy custom objects.
			},
			Name: config.Name,
		}

		resourceSetV2, err = v2.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV3 *framework.ResourceSet
	{
		c := v3.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			HandledVersionBundles: []string{
				"1.1.0",
			},
			Name: config.Name,
		}

		resourceSetV3, err = v3.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV4 *framework.ResourceSet
	{
		c := v4.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			HandledVersionBundles: []string{
				"1.2.0",
			},
			ProjectName: config.Name,
		}

		resourceSetV4, err = v4.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV5 *framework.ResourceSet
	{
		c := v5.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			HandledVersionBundles: []string{
				"2.0.0",
			},
			ProjectName: config.Name,
		}

		resourceSetV5, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{}

		c.ResourceSets = []*framework.ResourceSet{
			resourceSetV2,
			resourceSetV3,
			resourceSetV4,
			resourceSetV5,
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var crdFramework *framework.Framework
	{
		c := framework.Config{}

		c.CRD = v1alpha1.NewKVMConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.Logger = config.Logger
		c.ResourceRouter = resourceRouter

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}
