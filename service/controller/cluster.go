package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v11"
	"github.com/giantswarm/kvm-operator/service/controller/v12"
	v12cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v12/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v13"
	v13cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v13/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v14"
	v14cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v14/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch1"
	v14patch1cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v14patch1/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch2"
	v14patch2cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v14patch2/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v15"
	v15cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v15/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v16"
	v16cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v16/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v17"
	v17cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v17/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v2"
	"github.com/giantswarm/kvm-operator/service/controller/v4"
)

type ClusterConfig struct {
	CertsSearcher certs.Interface
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	K8sExtClient  apiextensionsclient.Interface
	Logger        micrologger.Logger

	CRDLabelSelector   string
	GuestUpdateEnabled bool
	OIDC               ClusterConfigOIDC
	ProjectName        string
	SSOPublicKey       string
}

// ClusterConfigOIDC represents the configuration of the OIDC authorization
// provider.
type ClusterConfigOIDC struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

func (c ClusterConfig) newInformerListOptions() metav1.ListOptions {
	listOptions := metav1.ListOptions{
		LabelSelector: c.CRDLabelSelector,
	}

	return listOptions
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
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
			Logger:  config.Logger,
			Watcher: config.G8sClient.ProviderV1alpha1().KVMConfigs(""),

			ListOptions:  config.newInformerListOptions(),
			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV2 *controller.ResourceSet
	{
		c := v2.ResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
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

	var resourceSetV4 *controller.ResourceSet
	{
		c := v4.ResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
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

	var resourceSetV11 *controller.ResourceSet
	{
		c := v11.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
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

	var resourceSetV12 *controller.ResourceSet
	{
		c := v12.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v12cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
		}

		resourceSetV12, err = v12.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV13 *controller.ResourceSet
	{
		c := v13.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v13cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV13, err = v13.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14 *controller.ResourceSet
	{
		c := v14.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v14cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV14, err = v14.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch1 *controller.ResourceSet
	{
		c := v14patch1.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v14patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV14Patch1, err = v14patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch2 *controller.ResourceSet
	{
		c := v14patch2.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v14patch2cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV14Patch2, err = v14patch2.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV15 *controller.ResourceSet
	{
		c := v15.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v15cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV15, err = v15.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV16 *controller.ResourceSet
	{
		c := v16.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v16cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV16, err = v16.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV17 *controller.ResourceSet
	{
		c := v17.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			OIDC: v17cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV17, err = v17.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewKVMConfigCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV2,
				resourceSetV4,
				resourceSetV11,
				resourceSetV12,
				resourceSetV13,
				resourceSetV14,
				resourceSetV14Patch1,
				resourceSetV14Patch2,
				resourceSetV15,
				resourceSetV16,
				resourceSetV17,
			},
			RESTClient: config.G8sClient.ProviderV1alpha1().RESTClient(),

			Name: config.ProjectName,
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: operatorkitController,
	}

	return c, nil
}
