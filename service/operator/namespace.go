package operator

import (
	"fmt"
	"time"

	"github.com/giantswarm/clusterspec"

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

func getNamespaceNameForCluster(cluster clusterspec.Cluster) string {
	// FIXME: I removed the prefix to make work our first version
	//return fmt.Sprintf("cluster-%v", cluster.Name)

	return fmt.Sprintf("%v", cluster.Name)
}

func (s *Service) createClusterNamespace(cluster clusterspec.Cluster) error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("create").Inc()

	s.logger.Log("debug", fmt.Sprintf("creating namespace for cluster '%s'", cluster.Name))

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
	if _, err = s.kubernetesClient.Core().Namespaces().Create(&namespace); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	if errors.IsAlreadyExists(err) {
		s.logger.Log("debug", "namespace already exists")
	} else {
		s.logger.Log("debug", "namespace created")
	}

	namespaceActionTime.WithLabelValues("create").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}

func (s *Service) deleteClusterNamespace(cluster clusterspec.Cluster) error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("delete").Inc()

	s.logger.Log("debug", fmt.Sprintf("deleting namespace for cluster '%s'", cluster.Name))

	namespaceName := getNamespaceNameForCluster(cluster)

	var err error
	if err = s.kubernetesClient.Core().Namespaces().Delete(namespaceName, v1.NewDeleteOptions(0)); err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		s.logger.Log("debug", "namespace already deleted")
	} else {
		s.logger.Log("debug", "namespace deleted")
	}

	namespaceActionTime.WithLabelValues("delete").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
