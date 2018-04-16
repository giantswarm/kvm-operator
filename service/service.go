package service

import (
	"sync"

	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/kvm-operator/service/controller"
	"github.com/giantswarm/kvm-operator/service/healthz"
)

type Config struct {
	Logger micrologger.Logger

	Description string
	Flag        *flag.Flag
	GitCommit   string
	Name        string
	Source      string
	Viper       *viper.Viper
}

type Service struct {
	Healthz          *healthz.Service
	ClusterFramework *framework.Framework
	DrainerFramework *framework.Framework
	Version          *version.Service

	bootOnce sync.Once
}

func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:   config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster: config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			TLS: k8srestconfig.TLSClientConfig{
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

	g8sClient, err := versioned.NewForConfig(restConfig)
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

	var healthzService *healthz.Service
	{
		healthzConfig := healthz.DefaultConfig()

		healthzConfig.K8sClient = k8sClient
		healthzConfig.Logger = config.Logger

		healthzService, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterFramework *framework.Framework
	{
		c := controller.ClusterFrameworkConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			GuestUpdateEnabled: config.Viper.GetBool(config.Flag.Service.Guest.Update.Enabled),
			ProjectName:        config.Name,
		}

		clusterFramework, err = controller.NewClusterFramework(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerFramework *framework.Framework
	{
		c := controller.DrainerFrameworkConfig{
			G8sClient: g8sClient,
			K8sClient: k8sClient,
			Logger:    config.Logger,

			ProjectName: config.Name,
		}

		drainerFramework, err = controller.NewDrainerFramework(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Description = config.Description
		versionConfig.GitCommit = config.GitCommit
		versionConfig.Name = config.Name
		versionConfig.Source = config.Source
		versionConfig.VersionBundles = NewVersionBundles()

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Healthz:          healthzService,
		ClusterFramework: clusterFramework,
		DrainerFramework: drainerFramework,
		Version:          versionService,

		bootOnce: sync.Once{},
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.ClusterFramework.Boot()
		go s.DrainerFramework.Boot()
	})
}
