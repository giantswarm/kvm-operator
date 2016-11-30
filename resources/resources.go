package resources

import (
	"bytes"
	"errors"
	"log"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus"

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

type ClusterObj interface {
	GenerateResources() ([]runtime.Object, error)
}

func ExecTemplate(t string, obj interface{}) (string, error) {
	var result bytes.Buffer

	tmpl, err := template.New("component").Parse(t)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&result, obj)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// computeResources returns a list of Kubernetes objects that define
// the desired state of the given cluster.
func ComputeResources(cluster *Cluster) ([]runtime.Object, error) {
	if cluster.Spec.ClusterID == "" {
		return nil, errors.New("cluster ID must not be empty")
	}
	if cluster.Spec.Replicas == int32(0) {
		return nil, errors.New("replicas must not be empty")
	}

	start := time.Now()
	computeResourcesTotal.WithLabelValues(cluster.Name).Inc()

	log.Println("started computing desired resources for cluster:", cluster.Name)

	objects := []runtime.Object{}

	configMap := &configMap{
		Cluster: *cluster,
	}
	configMapComponents, err := configMap.GenerateResources()
	objects = append(objects, configMapComponents...)

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

	// FlannelClient for the worker
	flannelComponents, err = flannelClient.GenerateResources()
	if err != nil {
		log.Println("generate resource flannelComponents error %v", err)
	}
	objects = append(objects, flannelComponents...)

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
