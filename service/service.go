// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	k8sutil "github.com/giantswarm/operatorkit/client/k8s"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/spf13/viper"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/kvm-operator/service/healthz"
	"github.com/giantswarm/kvm-operator/service/operator"
	cloudconfigresource "github.com/giantswarm/kvm-operator/service/resource/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/resource/legacy"
	masterresource "github.com/giantswarm/kvm-operator/service/resource/master"
	namespaceresource "github.com/giantswarm/kvm-operator/service/resource/namespace"
	workerresource "github.com/giantswarm/kvm-operator/service/resource/worker"
)

const (
	ListAPIEndpoint  = "/apis/cluster.giantswarm.io/v1/kvms"
	WatchAPIEndpoint = "/apis/cluster.giantswarm.io/v1/watch/kvms"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	Name        string
	Source      string
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Settings.
		Flag:  nil,
		Viper: nil,

		Description: "",
		GitCommit:   "",
		Name:        "",
		Source:      "",
	}
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating kvm-operator with config: %#v", config))

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sutil.Config{
			Logger: config.Logger,

			TLS: k8sutil.TLSClientConfig{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
			Address:   config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster: config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
		}

		k8sClient, err = k8sutil.NewClient(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.New(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudConfigResource legacy.Resource
	{
		cloudConfigConfig := cloudconfigresource.DefaultConfig()

		cloudConfigConfig.CertWatcher = certWatcher
		cloudConfigConfig.Logger = config.Logger

		cloudConfigResource, err = cloudconfigresource.New(cloudConfigConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var masterResource legacy.Resource
	{
		masterConfig := masterresource.DefaultConfig()

		masterConfig.Logger = config.Logger

		masterResource, err = masterresource.New(masterConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource legacy.Resource
	{
		namespaceConfig := namespaceresource.DefaultConfig()

		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespaceresource.New(namespaceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerResource legacy.Resource
	{
		workerConfig := workerresource.DefaultConfig()

		workerConfig.Logger = config.Logger

		workerResource, err = workerresource.New(workerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorFramework *framework.Framework
	{
		frameworkConfig := framework.DefaultConfig()

		frameworkConfig.Logger = config.Logger

		operatorFramework, err = framework.New(frameworkConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyResource *legacy.Reconciler
	{
		newConfig := legacy.DefaultConfig()

		// Dependencies.
		newConfig.K8sClient = k8sClient
		newConfig.Logger = config.Logger

		// Settings.
		newConfig.Resources = []legacy.Resource{
			// Note that the namespace resource is special since the creation of the
			// cluster namespace has to be done before any other resource can be
			// created inside of it. The current reconciliation is synchronous and
			// processes resources in a series. This is why the namespace resource has
			// to be registered first.
			namespaceResource,
			// Note that the cloud config resource is special since the creation of
			// configmaps has to be done before any pod can make use of it. The
			// current reconciliation is synchronous and processes resources in a
			// series. This is why the cloud config resource has to be registered
			// second.
			cloudConfigResource,
			masterResource,
			workerResource,
		}

		legacyResource, err = legacy.New(newConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
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

	var operatorService *operator.Service
	{
		operatorConfig := operator.DefaultConfig()

		operatorConfig.K8sClient = k8sClient
		operatorConfig.Logger = config.Logger
		operatorConfig.OperatorFramework = operatorFramework
		operatorConfig.Resources = []framework.Resource{
			legacyResource,
		}

		operatorService, err = operator.New(operatorConfig)
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

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Dependencies.
		Healthz:  healthzService,
		Operator: operatorService,
		Version:  versionService,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	Healthz  *healthz.Service
	Operator *operator.Service
	Version  *version.Service

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Operator.Boot()
	})
}
