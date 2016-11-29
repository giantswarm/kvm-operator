package controller

import (
	"fmt"
	"log"
	"time"

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

type Namespace interface {
	ClusterObj
}

type namespace struct {
	ClusterConfig
}


func init() {
	prometheus.MustRegister(namespaceActionTotal)
	prometheus.MustRegister(namespaceActionTime)
}

func (c *namespace) GetNamespace() string {
	return fmt.Sprintf("cluster-%v", c.ClusterID)
}

func (c *namespace) Create() error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("create").Inc()

	log.Println("creating namespace for cluster:", c.ClusterID)

	namespace := v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: c.GetNamespace(),
			Labels: map[string]string{
				"cluster":  c.ClusterID,
				"customer": c.ClusterID,
			},
		},
	}

	var err error
	if _, err = c.KubernetesClient.Core().Namespaces().Create(&namespace); err != nil && !errors.IsAlreadyExists(err) {
		return maskAny(err)
	}
	if errors.IsAlreadyExists(err) {
		log.Println("namespace already exists")
	} else {
		log.Println("namespace created")
	}

	namespaceActionTime.WithLabelValues("create").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}

func (c *namespace) Delete() error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("delete").Inc()

	log.Println("deleting namespace for cluster:", c.ClusterID)

	namespaceName := c.GetNamespace()

	var err error
	if err = c.KubernetesClient.Core().Namespaces().Delete(namespaceName, v1.NewDeleteOptions(0)); err != nil && !errors.IsNotFound(err) {
		return maskAny(err)
	}
	if errors.IsNotFound(err) {
		log.Println("namespace already deleted")
	} else {
		log.Println("namespace deleted")
	}

	namespaceActionTime.WithLabelValues("delete").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
