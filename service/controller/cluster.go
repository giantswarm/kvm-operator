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
	"github.com/giantswarm/tenantcluster"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	v20 "github.com/giantswarm/kvm-operator/service/controller/v20"
	v20cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v20/cloudconfig"
	v21 "github.com/giantswarm/kvm-operator/service/controller/v21"
	v21cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v21/cloudconfig"
	v22 "github.com/giantswarm/kvm-operator/service/controller/v22"
	v22cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v22/cloudconfig"
	v23 "github.com/giantswarm/kvm-operator/service/controller/v23"
	v23cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v23/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v23patch1"
	v23patch1cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v23patch1/cloudconfig"
	v24 "github.com/giantswarm/kvm-operator/service/controller/v24"
	v24cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v24/cloudconfig"
	v25 "github.com/giantswarm/kvm-operator/service/controller/v25"
	v25cloudconfig "github.com/giantswarm/kvm-operator/service/controller/v25/cloudconfig"
)

type ClusterConfig struct {
	CertsSearcher certs.Interface
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	K8sExtClient  apiextensionsclient.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface

	CRDLabelSelector   string
	DNSServers         string
	GuestUpdateEnabled bool
	IgnitionPath       string
	NTPServers         string
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
		c := randomkeys.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		randomkeysSearcher, err = randomkeys.NewSearcher(c)
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

	var resourceSetV20 *controller.ResourceSet
	{
		c := v20.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			ProjectName:        config.ProjectName,
			OIDC: v20cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV20, err = v20.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV21 *controller.ResourceSet
	{
		c := v21.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			ProjectName:        config.ProjectName,
			OIDC: v21cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV21, err = v21.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV22 *controller.ResourceSet
	{
		c := v22.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			ProjectName:        config.ProjectName,
			OIDC: v22cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV22, err = v22.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23 *controller.ResourceSet
	{
		c := v23.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			ProjectName:        config.ProjectName,
			OIDC: v23cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV23, err = v23.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23patch1 *controller.ResourceSet
	{
		c := v23patch1.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			ProjectName:        config.ProjectName,
			OIDC: v23patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV23patch1, err = v23patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV24 *controller.ResourceSet
	{
		c := v24.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			NTPServers:         config.NTPServers,
			ProjectName:        config.ProjectName,
			OIDC: v24cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV24, err = v24.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV25 *controller.ResourceSet
	{
		c := v25.ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.G8sClient,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			TenantCluster:      config.TenantCluster,

			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			NTPServers:         config.NTPServers,
			ProjectName:        config.ProjectName,
			OIDC: v25cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSetV25, err = v25.NewClusterResourceSet(c)
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
				resourceSetV20,
				resourceSetV21,
				resourceSetV22,
				resourceSetV23,
				resourceSetV23patch1,
				resourceSetV24,
				resourceSetV25,
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
