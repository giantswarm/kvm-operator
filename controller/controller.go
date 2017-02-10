package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/giantswarm/kvm-operator/resources"
	"github.com/giantswarm/clusterspec"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	ClusterListAPIEndpoint  = "/apis/giantswarm.io/v1/clusters"
	ClusterWatchAPIEndpoint = "/apis/giantswarm.io/v1/watch/clusters"
)

type Controller interface {
	Start()
}

type controller struct {
	clientset *kubernetes.Clientset
	config    Config
}

type Config struct {
	InCluster bool

	APIServer    string
	CAFilePath   string
	CertFilePath string
	KeyFilePath  string

	ListenAddress string
}

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

func New(config Config) Controller {
	var kubernetesConfig *rest.Config
	if config.InCluster {
		log.Println("using in cluster config")
		var err error
		kubernetesConfig, err = rest.InClusterConfig()
		if err != nil {
			panic(fmt.Sprintf("Could not create in cluster config: %v\n", err.Error()))
		}

		log.Println("using explicit api server")
		if config.APIServer != "" {
			kubernetesConfig.Host = config.APIServer
		}
	} else {
		log.Println("using external config")
		kubernetesConfig = &rest.Config{
			Host: config.APIServer,
			TLSClientConfig: rest.TLSClientConfig{
				CAFile:   config.CAFilePath,
				CertFile: config.CertFilePath,
				KeyFile:  config.KeyFilePath,
			},
		}
	}

	clientset := kubernetes.NewForConfigOrDie(kubernetesConfig)

	return &controller{
		clientset: clientset,
		config:    config,
	}
}

func (c *controller) newClusterListWatch() *cache.ListWatch {
	client := c.clientset.Core().RESTClient()

	listWatch := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			start := time.Now()
			clusterAPIActionTotal.WithLabelValues("list").Inc()

			req := client.Get().AbsPath(ClusterListAPIEndpoint)
			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}

			var c clusterspec.ClusterList
			if err := json.Unmarshal(b, &c); err != nil {
				return nil, err
			}

			clusterAPIActionTime.WithLabelValues("list").Set(float64(time.Since(start) / time.Millisecond))

			return &c, nil
		},

		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			start := time.Now()
			clusterAPIActionTotal.WithLabelValues("watch").Inc()

			req := client.Get().AbsPath(ClusterWatchAPIEndpoint)
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

// Start starts the server.
func (c *controller) Start() {
	go c.startServer()

	if err := c.createClusterResource(); err != nil {
		log.Fatalln("could not create cluster resource:", err)
	}

	_, clusterInformer := cache.NewInformer(
		c.newClusterListWatch(),
		&clusterspec.Cluster{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				start := time.Now()
				clusterEventHandleTotal.WithLabelValues("added").Inc()

				cluster := obj.(*clusterspec.Cluster)
				log.Printf("cluster '%v' added", cluster.Name)

				if err := c.createClusterNamespace(*cluster); err != nil {
					log.Println("could not create cluster namespace:", err)
				}

				// Given a cluster, determine the desired state,
				// in terms of resources that should exist in Kubernetes.
				resources, err := resources.ComputeResources(cluster)
				if err != nil {
					log.Println("could not compute required resources for cluster:", err)
				}

				// Reconcile the state of resources in Kubernetes with the desired state of resources we just computed.
				if err := c.reconcileResourceState(getNamespaceNameForCluster(*cluster), resources); err != nil {
					log.Println("could not reconcile resource state:", err)
				}

				clusterEventHandleTime.WithLabelValues("added").Set(float64(time.Since(start) / time.Millisecond))
			},
			DeleteFunc: func(obj interface{}) {
				start := time.Now()
				clusterEventHandleTotal.WithLabelValues("deleted").Inc()

				cluster := obj.(*clusterspec.Cluster)
				log.Printf("cluster '%v' deleted", cluster.Name)

				if err := c.deleteClusterNamespace(*cluster); err != nil {
					log.Println("could not delete cluster namespace:", err)
				}

				clusterEventHandleTime.WithLabelValues("deleted").Set(float64(time.Since(start) / time.Millisecond))
			},
		},
	)

	log.Println("starting watch")
	clusterInformer.Run(nil)
}
