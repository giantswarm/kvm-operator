package kvmconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/keyv3"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/resourcesv2"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/resourcesv3"
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
	var err error

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

	var resourcesV2 []framework.Resource
	{
		c := kvmconfigv2.ResourcesConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			Name: config.Name,
		}

		resourcesV2, err = kvmconfigv2.NewResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourcesV3 []framework.Resource
	{
		c := kvmconfigv3.ResourcesConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			Name: config.Name,
		}

		resourcesV3, err = kvmconfigv3.NewResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// versionedResources is a map of VersionBundle.Version to a resources
	// which handle it. The same set of resources may handle multiple
	// version bundles.
	versionedResources := map[string][]framework.Resource{
		"1.1.0": resourcesV3,
		"1.0.0": resourcesV2,
		"0.1.0": resourcesV2,
		"":      resourcesV2,
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

	// TODO route initCtx func together with resources.
	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		{
			customObject, err := keyv3.ToCustomObject(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			versionBundleVersion := keyv3.VersionBundleVersion(customObject)

			if config.GuestUpdateEnabled && versionBundleVersion >= "1.1.0" {
				updateallowedcontext.SetUpdateAllowed(ctx)
			}
		}

		return ctx, nil
	}

	var crdFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.CRD = v1alpha1.NewKVMConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.InitCtxFunc = initCtxFunc
		c.Logger = config.Logger
		c.ResourceRouter = newResourceRouter(versionedResources)

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}
