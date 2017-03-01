package resources

import (
	"errors"
	"log"
	"time"

	"github.com/giantswarm/clusterspec"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/runtime"
)

const (
	MasterReplicas int32 = 1
)

var (
	computeResourcesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "compute_resources_total",
			Help: "Number of times we have computed resources for a cluster",
		},
		[]string{"cluster"},
	)
	computeResourceTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "compute_resource_milliseconds",
			Help: "Time taken to handle computing resources for a cluster, in milliseconds",
		},
		[]string{"cluster"},
	)
)

func init() {
	prometheus.MustRegister(computeResourcesTotal)
	prometheus.MustRegister(computeResourceTime)
}

type ClusterObj interface {
	GenerateResources() ([]runtime.Object, error)
}

// computeResources returns a list of Kubernetes objects that define
// the desired state of the given cluster.
func ComputeResources(customObject clusterspec.Cluster) ([]runtime.Object, error) {
	clusterID := ClusterID(customObject)

	if clusterID == "" {
		return nil, errors.New("cluster ID must not be empty")
	}
	if customObject.Spec.Worker.Replicas == int32(0) {
		return nil, errors.New("worker replicas must not be empty")
	}

	start := time.Now()
	computeResourcesTotal.WithLabelValues(clusterID).Inc()

	log.Println("started computing desired resources for cluster:", clusterID)

	objects := []runtime.Object{}

	flannelClient := &flannelClient{
		Cluster: customObject,
	}
	flannelComponents, err := flannelClient.GenerateResources()
	if err != nil {
		log.Println("generate resource flannelComponents error %v", err)
	}
	objects = append(objects, flannelComponents...)

	master := &master{
		Cluster: customObject,
	}
	masterComponents, err := master.GenerateResources()
	if err != nil {
		log.Println("generate resource masterComponents error %v", err)
	}
	objects = append(objects, masterComponents...)

	worker := &worker{
		Cluster: customObject,
	}
	workerComponents, err := worker.GenerateResources()
	if err != nil {
		log.Println("generate resource workerComponents error %v", err)
	}
	objects = append(objects, workerComponents...)

	log.Println("finished computing desired resources for cluster:", clusterID)

	computeResourceTime.WithLabelValues(clusterID).Set(float64(time.Since(start) / time.Millisecond))

	return objects, nil
}
