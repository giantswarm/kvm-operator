package resources

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/giantswarm/kvmtpr"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/pkg/runtime"
)

const (
	MasterReplicas int = 1
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
func ComputeResources(customObject kvmtpr.CustomObject) ([]runtime.Object, error) {
	clusterID := ClusterID(customObject)

	if clusterID == "" {
		return nil, errors.New("cluster ID must not be empty")
	}
	if len(customObject.Spec.Cluster.Workers) == 0 {
		return nil, errors.New("worker replicas must not be empty")
	}

	start := time.Now()
	computeResourcesTotal.WithLabelValues(clusterID).Inc()

	fmt.Printf("started computing desired resources for cluster '%s'\n", clusterID)

	objects := []runtime.Object{}

	flannelClient := &flannelClient{
		CustomObject: customObject,
	}
	flannelComponents, err := flannelClient.GenerateResources()
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	objects = append(objects, flannelComponents...)

	master := &master{
		CustomObject: customObject,
	}
	masterComponents, err := master.GenerateResources()
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	objects = append(objects, masterComponents...)

	worker := &worker{
		CustomObject: customObject,
	}
	workerComponents, err := worker.GenerateResources()
	if err != nil {
		log.Printf("generate resource workerComponents error: %v\n", err)
	}
	objects = append(objects, workerComponents...)

	fmt.Printf("finished computing desired resources for cluster '%s'\n", clusterID)

	computeResourceTime.WithLabelValues(clusterID).Set(float64(time.Since(start) / time.Millisecond))

	return objects, nil
}
