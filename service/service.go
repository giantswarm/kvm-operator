package service

import (
	"context"
	"sync"
	"time"

	v1alpha12 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v5/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/statusresource/v3"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

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
	machineController       *controller.Machine
	transitionController    *controller.Transition
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

	if config.Viper.GetString(config.Flag.Service.Registry.DockerhubToken) == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag.Service.Registry.DockerhubToken must not be empty", config)
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
				v1alpha2.AddToScheme,
				capiv1alpha3.AddToScheme,
				v1alpha12.AddToScheme,
				releasev1alpha1.AddToScheme,
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

			ClusterRoleGeneral: config.Viper.GetString(config.Flag.Service.RBAC.ClusterRole.General),
			ClusterRolePSP:     config.Viper.GetString(config.Flag.Service.RBAC.ClusterRole.PSP),
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
		}

		drainerController, err = controller.NewDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineController *controller.Machine
	{
		c := controller.MachineConfig{
			CertsSearcher: certsSearcher,
			K8sClient:     k8sClient,
			Logger:        config.Logger,
			TenantCluster: tenantCluster,

			DNSServers:   config.Viper.GetString(config.Flag.Service.Installation.DNS.Servers),
			IgnitionPath: config.Viper.GetString(config.Flag.Service.Tenant.Ignition.Path),
			NTPServers:   config.Viper.GetString(config.Flag.Service.Installation.NTP.Servers),
			SSOPublicKey: config.Viper.GetString(config.Flag.Service.Tenant.SSH.SSOPublicKey),

			DockerhubToken:  config.Viper.GetString(config.Flag.Service.Registry.DockerhubToken),
			RegistryDomain:  config.Viper.GetString(config.Flag.Service.Registry.Domain),
			RegistryMirrors: config.Viper.GetStringSlice(config.Flag.Service.Registry.Mirrors),
		}

		machineController, err = controller.NewMachine(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var transitionController *controller.Transition
	{
		c := controller.TransitionConfig{
			K8sClient:     k8sClient,
			Logger:        config.Logger,
			TenantCluster: tenantCluster,
		}

		transitionController, err = controller.NewTransition(c)
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
		machineController:       machineController,
		transitionController:    transitionController,
		statusResourceCollector: statusResourceCollector,
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		ctx := context.Background()

		go func() {
			err := s.statusResourceCollector.Boot(ctx)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}()

		go s.clusterController.Boot(ctx)
		go s.deleterController.Boot(ctx)
		go s.drainerController.Boot(ctx)
		go s.machineController.Boot(ctx)
		go s.transitionController.Boot(ctx)
	})
}
