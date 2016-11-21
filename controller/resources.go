package controller

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
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

// computeResources returns a list of Kubernetes objects that define
// the desired state of the given cluster.
func (c *controller) computeResources(cluster *Cluster) ([]runtime.Object, error) {
	start := time.Now()
	computeResourcesTotal.WithLabelValues(cluster.Name).Inc()

	log.Println("started computing desired resources for cluster:", cluster.Name)

	objects := []runtime.Object{}

	exampleService := &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "example-service",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "http",
					Port: int32(8000),
				},
			},
		},
	}

	objects = append(objects, exampleService)

	log.Println("finished computing desired resources for cluster:", cluster.Name)

	computeResourceTime.WithLabelValues(cluster.Name).Set(float64(time.Since(start) / time.Millisecond))

	return objects, nil
}
