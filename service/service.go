// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"fmt"
	"net/url"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"

	"github.com/giantswarm/kvm-operator/service/healthz"
	"github.com/giantswarm/kvm-operator/service/operator"
	"github.com/giantswarm/kvm-operator/service/version"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	KubernetesAddress    string
	KubernetesInCluster  bool
	KubernetesTLSCAFile  string
	KubernetesTLSCrtFile string
	KubernetesTLSKeyFile string

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
		KubernetesAddress:    "",
		KubernetesInCluster:  false,
		KubernetesTLSCAFile:  "",
		KubernetesTLSCrtFile: "",
		KubernetesTLSKeyFile: "",

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

		if config.KubernetesInCluster {
			config.Logger.Log("debug", "creating in-cluster config")
			restConfig, err = rest.InClusterConfig()
			if err != nil {
				return nil, microerror.MaskAny(err)
			}

			if config.KubernetesAddress != "" {
				config.Logger.Log("debug", "using explicit api server")
				restConfig.Host = config.KubernetesAddress
			}
		} else {
			if config.KubernetesAddress == "" {
				return nil, microerror.MaskAnyf(invalidConfigError, "kubernetes address must not be empty")
			}

			config.Logger.Log("debug", "creating out-cluster config")

			// Kubernetes listen URL.
			u, err := url.Parse(config.KubernetesAddress)
			if err != nil {
				return nil, microerror.MaskAny(err)
			}

			restConfig = &rest.Config{
				Host: u.String(),
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   config.KubernetesTLSCAFile,
					CertFile: config.KubernetesTLSCrtFile,
					KeyFile:  config.KubernetesTLSKeyFile,
				},
			}
		}

		kubernetesClient, err = kubernetes.NewForConfig(restConfig)
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

		operatorConfig.Logger = config.Logger
		operatorConfig.KubernetesClient = kubernetesClient

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
