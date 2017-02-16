package healthz

import (
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/giantswarm/kvm-operator/service/operator"
)

var (
	healthCheckRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_check_request_total",
			Help: "Number of health check requests",
		},
		[]string{"success"},
	)
	healthCheckRequestTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "health_check_request_milliseconds",
		Help: "Time taken to respond to health check, in milliseconds",
	})
)

func init() {
	prometheus.MustRegister(healthCheckRequests)
	prometheus.MustRegister(healthCheckRequestTime)
}

// Config represents the configuration used to create a version service.
type Config struct {
	// Dependencies.
	KubernetesClient *kubernetes.Clientset
	Logger           micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		KubernetesClient: nil,
		Logger:           nil,
	}
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.KubernetesClient == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "kubernetes client must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		kubernetesClient: config.KubernetesClient,
		logger:           config.Logger,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	kubernetesClient *kubernetes.Clientset
	logger           micrologger.Logger

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Check(ctx context.Context, request Request) (*Response, error) {
	start := time.Now()
	defer func() {
		healthCheckRequestTime.Set(float64(time.Since(start) / time.Millisecond))
	}()

	_, err := s.kubernetesClient.Extensions().ThirdPartyResources().Get(operator.ClusterThirdPartyResourceName)
	if err != nil {
		healthCheckRequests.WithLabelValues("failed").Inc()
		return nil, microerror.MaskAny(err)
	}

	healthCheckRequests.WithLabelValues("successful").Inc()

	return DefaultResponse(), nil
}
