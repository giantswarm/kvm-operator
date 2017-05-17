package operator

import (
	"sync"

	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	k8sreconciler "github.com/giantswarm/kvm-operator/service/reconciler/k8s"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	KubernetesClient *kubernetes.Clientset
	Logger           micrologger.Logger
	Reconciler       *k8sreconciler.Service
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		KubernetesClient: nil,
		Logger:           nil,
		Reconciler:       nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.KubernetesClient == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Kubernetes client must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Reconciler == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Reconciler must not be empty")
	}

	newService := &Service{
		// Dependencies.
		kubernetesClient: config.KubernetesClient,
		logger:           config.Logger,
		reconciler:       config.Reconciler,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the service.
type Service struct {
	// Dependencies.
	kubernetesClient *kubernetes.Clientset
	logger           micrologger.Logger
	reconciler       *k8sreconciler.Service

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		if err := s.createClusterResource(); err != nil {
			panic(err)
		}

		_, clusterInformer := cache.NewInformer(
			s.reconciler.GetListWatch(),
			&kvmtpr.CustomObject{},
			0,
			cache.ResourceEventHandlerFuncs{
				AddFunc:    s.reconciler.GetAddFunc(),
				DeleteFunc: s.reconciler.GetDeleteFunc(),
			},
		)

		s.logger.Log("debug", "starting list/watch")
		clusterInformer.Run(nil)
	})
}
