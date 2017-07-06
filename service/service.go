// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"fmt"
	"net/url"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/certificatetpr"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"github.com/spf13/viper"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/kvm-operator/service/healthz"
	"github.com/giantswarm/kvm-operator/service/operator"
	k8sreconciler "github.com/giantswarm/kvm-operator/service/reconciler/k8s"
	cloudconfigresource "github.com/giantswarm/kvm-operator/service/resource/cloudconfig"
	flannelservice "github.com/giantswarm/kvm-operator/service/resource/flannel"
	masterresource "github.com/giantswarm/kvm-operator/service/resource/master"
	namespaceresource "github.com/giantswarm/kvm-operator/service/resource/namespace"
	workerresource "github.com/giantswarm/kvm-operator/service/resource/worker"
	"github.com/giantswarm/kvm-operator/service/version"
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
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating kvm-operator with config: %#v", config))

	var err error

	var kubernetesClient *kubernetes.Clientset
	{
		var restConfig *rest.Config
		address := config.Viper.GetString(config.Flag.Service.Kubernetes.Address)

		if config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster) {
			config.Logger.Log("debug", "creating in-cluster config")
			restConfig, err = rest.InClusterConfig()
			if err != nil {
				return nil, microerror.MaskAny(err)
			}

			if address != "" {
				config.Logger.Log("debug", "using explicit api server")
				restConfig.Host = address
			}
		} else {
			if address == "" {
				return nil, microerror.MaskAnyf(invalidConfigError, "kubernetes address must not be empty")
			}

			config.Logger.Log("debug", "creating out-cluster config")

			// Kubernetes listen URL.
			u, err := url.Parse(address)
			if err != nil {
				return nil, microerror.MaskAny(err)
			}

			restConfig = &rest.Config{
				Host: u.String(),
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CaFile),
					CertFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
					KeyFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
				},
			}
		}

		kubernetesClient, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultConfig()
		certConfig.K8sClient = kubernetesClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.New(certConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var cloudConfigResource k8sreconciler.Resource
	{
		cloudConfigConfig := cloudconfigresource.DefaultConfig()

		cloudConfigConfig.CertWatcher = certWatcher
		cloudConfigConfig.Logger = config.Logger

		cloudConfigResource, err = cloudconfigresource.New(cloudConfigConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var flannel *flannelservice.Service
	{
		flannelConfig := flannelservice.DefaultConfig()

		flannelConfig.Logger = config.Logger

		flannel, err = flannelservice.New(flannelConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var masterResource k8sreconciler.Resource
	{
		masterConfig := masterresource.DefaultConfig()

		masterConfig.Logger = config.Logger
		masterConfig.Flannel = flannel

		masterResource, err = masterresource.New(masterConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var namespaceResource k8sreconciler.Resource
	{
		namespaceConfig := namespaceresource.DefaultConfig()

		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespaceresource.New(namespaceConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var workerResource k8sreconciler.Resource
	{
		workerConfig := workerresource.DefaultConfig()

		workerConfig.Logger = config.Logger
		workerConfig.Flannel = flannel

		workerResource, err = workerresource.New(workerConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var newReconciler *k8sreconciler.Reconciler
	{
		newConfig := k8sreconciler.DefaultConfig()

		// Dependencies.
		newConfig.KubernetesClient = kubernetesClient
		newConfig.ListDecoder = &listDecoder{}
		newConfig.Logger = config.Logger

		// Settings.
		newConfig.ListEndpoint = ListAPIEndpoint
		newConfig.Resources = []k8sreconciler.Resource{
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
		newConfig.WatchEndpoint = WatchAPIEndpoint

		newReconciler, err = k8sreconciler.New(newConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var healthzService *healthz.Service
	{
		healthzConfig := healthz.DefaultConfig()

		healthzConfig.Logger = config.Logger
		healthzConfig.KubernetesClient = kubernetesClient

		healthzService, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var operatorService *operator.Service
	{
		operatorConfig := operator.DefaultConfig()

		operatorConfig.KubernetesClient = kubernetesClient
		operatorConfig.Logger = config.Logger
		operatorConfig.Reconciler = newReconciler

		operatorService, err = operator.New(operatorConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
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
			return nil, microerror.MaskAny(err)
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
