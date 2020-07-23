package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/randomkeys"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
)

type ClusterConfig struct {
	CertsSearcher certs.Interface
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface

	ClusterRoleGeneral string
	ClusterRolePSP     string
	CRDLabelSelector   string
	DNSServers         string
	GuestUpdateEnabled bool
	IgnitionPath       string
	NTPServers         string
	OIDC               ClusterConfigOIDC
	ProjectName        string
	RegistryDomain     string
	RegistryMirrors    []string
	SSOPublicKey       string
}

// ClusterConfigOIDC represents the configuration of the OIDC authorization
// provider.
type ClusterConfigOIDC struct {
	ClientID       string
	IssuerURL      string
	UsernameClaim  string
	UsernamePrefix string
	GroupsClaim    string
	GroupsPrefix   string
}

type Registry struct{}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	resourceSets, err := newClusterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.KVMConfig)
			},

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

func newClusterResourceSets(config ClusterConfig) ([]*controller.ResourceSet, error) {
	var err error

	var randomkeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomkeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSet *controller.ResourceSet
	{
		c := ClusterResourceSetConfig{
			CertsSearcher:      config.CertsSearcher,
			G8sClient:          config.K8sClient.G8sClient(),
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,
			RegistryDomain:     config.RegistryDomain,
			TenantCluster:      config.TenantCluster,

			ClusterRoleGeneral: config.ClusterRoleGeneral,
			ClusterRolePSP:     config.ClusterRolePSP,
			DNSServers:         config.DNSServers,
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			IgnitionPath:       config.IgnitionPath,
			NTPServers:         config.NTPServers,
			ProjectName:        config.ProjectName,
			OIDC: cloudconfig.OIDCConfig{
				ClientID:       config.OIDC.ClientID,
				IssuerURL:      config.OIDC.IssuerURL,
				UsernameClaim:  config.OIDC.UsernameClaim,
				UsernamePrefix: config.OIDC.UsernamePrefix,
				GroupsClaim:    config.OIDC.GroupsClaim,
				GroupsPrefix:   config.OIDC.GroupsPrefix,
			},
			SSOPublicKey: config.SSOPublicKey,
		}

		resourceSet, err = NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
