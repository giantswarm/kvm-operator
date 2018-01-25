package service

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/kvm-operator/service/keyv3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func newKVMConfigFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	clientSet, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = k8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certsSearcher certs.Interface
	{
		c := certs.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomkeysSearcher randomkeys.Interface
	{
		keyConfig := randomkeys.DefaultConfig()
		keyConfig.K8sClient = k8sClient
		keyConfig.Logger = config.Logger
		randomkeysSearcher, err = randomkeys.NewSearcher(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourcesV2 []framework.Resource
	{
		c := defaultKVMConfigResourcesV2Config()

		c.CertsSearcher = certsSearcher
		c.K8sClient = k8sClient
		c.Logger = config.Logger
		c.RandomkeysSearcher = randomkeysSearcher

		c.Name = config.Name

		resourcesV2, err = newKVMConfigResourcesV2(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourcesV3 []framework.Resource
	{
		c := defaultKVMConfigResourcesV3Config()

		c.CertsSearcher = certsSearcher
		c.K8sClient = k8sClient
		c.Logger = config.Logger
		c.RandomkeysSearcher = randomkeysSearcher

		c.Name = config.Name

		resourcesV3, err = newKVMConfigResourcesV3(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We provide a map of resource lists keyed by the version bundle version
	// to the resource router.
	versionedResources := map[string][]framework.Resource{
		"1.1.0": resourcesV3,
		"1.0.0": resourcesV2,
		"0.1.0": resourcesV2,
		"":      resourcesV2,
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = clientSet.ProviderV1alpha1().KVMConfigs("")

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

			updateEnabled := config.Viper.GetBool(config.Flag.Service.Guest.Update.Enabled)
			versionBundleVersion := keyv3.VersionBundleVersion(customObject)

			if updateEnabled && versionBundleVersion >= "1.1.0" {
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
		// TODO extend ResourceRouter with context init.
		c.ResourceRouter = newKVMConfigResourceRouter(versionedResources)

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}
