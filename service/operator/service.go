package operator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/giantswarm/clusterspec"
	"github.com/giantswarm/kvm-operator/resources"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ClusterListAPIEndpoint  = "/apis/giantswarm.io/v1/clusters"
	ClusterWatchAPIEndpoint = "/apis/giantswarm.io/v1/watch/clusters"
)

var (
	clusterAPIActionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cluster_api_total",
			Help: "Number of cluster api actions",
		},
		[]string{"action"},
	)
	clusterAPIActionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_api_milliseconds",
			Help: "Time taken to handle cluster api action, in milliseconds",
		},
		[]string{"action"},
	)

	clusterEventHandleTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cluster_event_handle_total",
			Help: "Number of cluster events handled",
		},
		[]string{"action"},
	)
	clusterEventHandleTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_resource_handle_milliseconds",
			Help: "Time taken to handle cluster event, in milliseconds",
		},
		[]string{"action"},
	)
)

func init() {
	prometheus.MustRegister(clusterAPIActionTotal)
	prometheus.MustRegister(clusterAPIActionTime)
	prometheus.MustRegister(clusterEventHandleTotal)
	prometheus.MustRegister(clusterEventHandleTime)
}

// Config represents the configuration used to create a operator service.
type Config struct {
	// Dependencies.
	KubernetesClient *kubernetes.Clientset
	Logger           micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new operator
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		KubernetesClient: nil,
		Logger:           nil,
	}
}

// New creates a new configured operator service.
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

// Service implements the operator service interface.
type Service struct {
	// Dependencies.
	kubernetesClient *kubernetes.Clientset
	logger           micrologger.Logger

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		if err := s.createClusterResource(); err != nil {
			panic(fmt.Sprintf("could not create cluster resource: %#v", err))
		}

		_, clusterInformer := cache.NewInformer(
			s.newClusterListWatch(),
			&clusterspec.Cluster{},
			0,
			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					start := time.Now()
					clusterEventHandleTotal.WithLabelValues("added").Inc()

					cluster := obj.(*clusterspec.Cluster)
					s.logger.Log("debug", fmt.Sprintf("cluster created '%s'", cluster.Name))

					err := s.createClusterNamespace(*cluster)
					if err != nil {
						s.logger.Log("error", fmt.Sprintf("could not create cluster namespace '%#v'", err))
					}

					// Given a cluster, determine the desired state,
					// in terms of resources that should exist in Kubernetes.
					resources, err := resources.ComputeResources(cluster)
					if err != nil {
						s.logger.Log("error", fmt.Sprintf("could not compute required resources for cluster '%#v'", err))
					}

					// Reconcile the state of resources in Kubernetes with the desired state of resources we just computed.
					err = s.reconcileResourceState(getNamespaceNameForCluster(*cluster), resources)
					if err != nil {
						s.logger.Log("error", fmt.Sprintf("could not reconcile resource state '%#v'", err))
					}

					clusterEventHandleTime.WithLabelValues("added").Set(float64(time.Since(start) / time.Millisecond))
				},
				DeleteFunc: func(obj interface{}) {
					start := time.Now()
					clusterEventHandleTotal.WithLabelValues("deleted").Inc()

					cluster := obj.(*clusterspec.Cluster)
					s.logger.Log("debug", fmt.Sprintf("cluster deleted '%s'", cluster.Name))

					err := s.deleteClusterNamespace(*cluster)
					if err != nil {
						s.logger.Log("error", fmt.Sprintf("could not delete cluster namespace '%#v'", err))
					}

					clusterEventHandleTime.WithLabelValues("deleted").Set(float64(time.Since(start) / time.Millisecond))
				},
			},
		)

		s.logger.Log("debug", "starting watch")
		clusterInformer.Run(nil)
	})
}

func (s *Service) newClusterListWatch() *cache.ListWatch {
	restClient := s.kubernetesClient.Core().RESTClient()

	listWatch := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			start := time.Now()
			clusterAPIActionTotal.WithLabelValues("list").Inc()

			req := restClient.Get().AbsPath(ClusterListAPIEndpoint)
			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}

			var s clusterspec.ClusterList
			if err := json.Unmarshal(b, &s); err != nil {
				return nil, err
			}

			clusterAPIActionTime.WithLabelValues("list").Set(float64(time.Since(start) / time.Millisecond))

			return &s, nil
		},

		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			start := time.Now()
			clusterAPIActionTotal.WithLabelValues("watch").Inc()

			req := restClient.Get().AbsPath(ClusterWatchAPIEndpoint)
			stream, err := req.Stream()
			if err != nil {
				return nil, err
			}

			watcher := watch.NewStreamWatcher(&clusterDecoder{
				decoder: json.NewDecoder(stream),
				close:   stream.Close,
			})

			clusterAPIActionTime.WithLabelValues("watch").Set(float64(time.Since(start) / time.Millisecond))

			return watcher, nil
		},
	}

	return listWatch
}
