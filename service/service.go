package service

import (
	"context"
	"sync"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/statusresource"
	"github.com/giantswarm/tenantcluster"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service

	bootOnce                sync.Once
	clusterController       *controller.Cluster
	deleterController       *controller.Deleter
	drainerController       *controller.Drainer
	statusResourceCollector *statusresource.CollectorSet
}

// New creates a new service with given configuration.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag must not be empty", config)
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Viper must not be empty", config)
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:    config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				v1alpha1.AddToScheme,
			},
			Logger: config.Logger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: k8sClient.K8sClient(),
			Logger:    config.Logger,

			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,

			CertID: certs.APICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Cluster
	{
		c := controller.ClusterConfig{
			CertsSearcher: certsSearcher,
			K8sClient:     k8sClient,
			Logger:        config.Logger,
			TenantCluster: tenantCluster,

			ClusterRoleGeneral: config.Viper.GetString(config.Flag.Service.RBAC.ClusterRole.General),
			ClusterRolePSP:     config.Viper.GetString(config.Flag.Service.RBAC.ClusterRole.PSP),
			CRDLabelSelector:   config.Viper.GetString(config.Flag.Service.CRD.LabelSelector),
			GuestUpdateEnabled: config.Viper.GetBool(config.Flag.Service.Tenant.Update.Enabled),
			ProjectName:        project.Name(),

			DNSServers:   config.Viper.GetString(config.Flag.Service.Installation.DNS.Servers),
			IgnitionPath: config.Viper.GetString(config.Flag.Service.Tenant.Ignition.Path),
			NTPServers:   config.Viper.GetString(config.Flag.Service.Installation.NTP.Servers),
			OIDC: controller.ClusterConfigOIDC{
				ClientID:      config.Viper.GetString(config.Flag.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.ClientID),
				IssuerURL:     config.Viper.GetString(config.Flag.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.IssuerURL),
				UsernameClaim: config.Viper.GetString(config.Flag.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim),
				GroupsClaim:   config.Viper.GetString(config.Flag.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim),
			},
			RegistryDomain: config.Viper.GetString(config.Flag.Service.RegistryDomain),
			SSOPublicKey:   config.Viper.GetString(config.Flag.Service.Tenant.SSH.SSOPublicKey),
		}

		clusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deleterController *controller.Deleter
	{
		c := controller.DeleterConfig{
			CertsSearcher: certsSearcher,
			K8sClient:     k8sClient,
			Logger:        config.Logger,
			TenantCluster: tenantCluster,

			CRDLabelSelector: config.Viper.GetString(config.Flag.Service.CRD.LabelSelector),
			ProjectName:      project.Name(),
		}

		deleterController, err = controller.NewDeleter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerController *controller.Drainer
	{
		c := controller.DrainerConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,

			CRDLabelSelector: config.Viper.GetString(config.Flag.Service.CRD.LabelSelector),
			ProjectName:      project.Name(),
		}

		drainerController, err = controller.NewDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResourceCollector *statusresource.CollectorSet
	{
		c := statusresource.CollectorSetConfig{
			Logger:  config.Logger,
			Watcher: k8sClient.G8sClient().ProviderV1alpha1().KVMConfigs("").Watch,
		}

		statusResourceCollector, err = statusresource.NewCollectorSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.Config{
			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: []versionbundle.Bundle{project.NewVersionBundle()},
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Version: versionService,

		bootOnce:                sync.Once{},
		clusterController:       clusterController,
		deleterController:       deleterController,
		drainerController:       drainerController,
		statusResourceCollector: statusResourceCollector,
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.statusResourceCollector.Boot(context.Background())

		go s.clusterController.Boot(context.Background())
		go s.deleterController.Boot(context.Background())
		go s.drainerController.Boot(context.Background())
	})
}
