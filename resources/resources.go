package resources

import (
	"errors"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/runtime"
)

const (
	GiantnetesConfigMapName string = "g8s-configmap"
	MasterReplicas          int32  = 1
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
func ComputeResources(cluster *Cluster) ([]runtime.Object, error) {
	if cluster.Spec.ClusterID == "" {
		return nil, errors.New("cluster ID must not be empty")
	}
	if cluster.Spec.WorkerReplicas == int32(0) {
		return nil, errors.New("worker replicas must not be empty")
	}

	start := time.Now()
	computeResourcesTotal.WithLabelValues(cluster.Name).Inc()

	log.Println("started computing desired resources for cluster:", cluster.Name)

	objects := []runtime.Object{}

	flannelClient := &flannelClient{
		Cluster: *cluster,
	}
	flannelComponents, err := flannelClient.GenerateResources()
	if err != nil {
		log.Println("generate resource flannelComponents error %v", err)
	}
	objects = append(objects, flannelComponents...)

	master := &master{
		Cluster: *cluster,
	}
	masterComponents, err := master.GenerateResources()
	if err != nil {
		log.Println("generate resource masterComponents error %v", err)
	}
	objects = append(objects, masterComponents...)

	worker := &worker{
		Cluster: *cluster,
	}
	workerComponents, err := worker.GenerateResources()
	if err != nil {
		log.Println("generate resource workerComponents error %v", err)
	}
	objects = append(objects, workerComponents...)

	log.Println("finished computing desired resources for cluster:", cluster.Name)

	computeResourceTime.WithLabelValues(cluster.Name).Set(float64(time.Since(start) / time.Millisecond))

	return objects, nil
}
