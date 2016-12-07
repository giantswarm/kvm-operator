package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/giantswarm/cluster-controller/resources"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

var (
	namespaceActionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "namespace_action_total",
			Help: "Number of namespace actions",
		},
		[]string{"action"},
	)
	namespaceActionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "namespace_action_milliseconds",
			Help: "Time taken to perform namespace action, in milliseconds",
		},
		[]string{"action"},
	)
)

func init() {
	prometheus.MustRegister(namespaceActionTotal)
	prometheus.MustRegister(namespaceActionTime)
}

func getNamespaceNameForCluster(cluster resources.Cluster) string {
	// FIXME: I removed the prefix to make work our first version
	//return fmt.Sprintf("cluster-%v", cluster.Name)

	return fmt.Sprintf("%v", cluster.Name)
}

func (c *controller) createClusterNamespace(cluster resources.Cluster) error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("create").Inc()

	log.Println("creating namespace for cluster:", cluster.Name)

	namespace := v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: getNamespaceNameForCluster(cluster),
			Labels: map[string]string{
				"cluster":  cluster.Name,
				"customer": cluster.Spec.Customer,
			},
		},
	}

	var err error
	if _, err = c.clientset.Core().Namespaces().Create(&namespace); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	if errors.IsAlreadyExists(err) {
		log.Println("namespace already exists")
	} else {
		log.Println("namespace created")
	}

	namespaceActionTime.WithLabelValues("create").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}

func (c *controller) deleteClusterNamespace(cluster resources.Cluster) error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("delete").Inc()

	log.Println("deleting namespace for cluster:", cluster.Name)

	namespaceName := getNamespaceNameForCluster(cluster)

	var err error
	if err = c.clientset.Core().Namespaces().Delete(namespaceName, v1.NewDeleteOptions(0)); err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		log.Println("namespace already deleted")
	} else {
		log.Println("namespace deleted")
	}

	namespaceActionTime.WithLabelValues("delete").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
