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

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v10"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v11"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v2"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v4"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v5"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v6"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v7"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v8"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v9"
)

type ClusterFrameworkConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestUpdateEnabled bool
	ProjectName        string
}

func NewClusterFramework(config ClusterFrameworkConfig) (*framework.Framework, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}

	var err error

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: config.K8sExtClient,
			Logger:       config.Logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

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
		c := informer.Config{
			Watcher: config.G8sClient.ProviderV1alpha1().KVMConfigs(""),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

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
			Name: config.ProjectName,
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
			Name:               config.ProjectName,
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
			ProjectName:        config.ProjectName,
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
			ProjectName:        config.ProjectName,
		}

		resourceSetV5, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV6 *framework.ResourceSet
	{
		c := v6.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		resourceSetV6, err = v6.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV7 *framework.ResourceSet
	{
		c := v7.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		resourceSetV7, err = v7.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV8 *framework.ResourceSet
	{
		c := v8.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		resourceSetV8, err = v8.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV9 *framework.ResourceSet
	{
		c := v9.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		resourceSetV9, err = v9.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV10 *framework.ResourceSet
	{
		c := v10.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		resourceSetV10, err = v10.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV11 *framework.ResourceSet
	{
		c := v11.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		resourceSetV11, err = v11.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{
			Logger: config.Logger,

			ResourceSets: []*framework.ResourceSet{
				resourceSetV2,
				resourceSetV3,
				resourceSetV4,
				resourceSetV5,
				resourceSetV6,
				resourceSetV7,
				resourceSetV8,
				resourceSetV9,
				resourceSetV10,
				resourceSetV11,
			},
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterFramework *framework.Framework
	{
		c := framework.Config{
			CRD:            v1alpha1.NewKVMConfigCRD(),
			CRDClient:      crdClient,
			Informer:       newInformer,
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,

			Name: config.ProjectName,
		}

		clusterFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return clusterFramework, nil
}
