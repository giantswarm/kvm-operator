package operator

import (
	"fmt"
	"time"

	"github.com/giantswarm/kvm-operator/resources"
	"github.com/giantswarm/kvmtpr"
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

func (s *Service) createNamespace(customObject kvmtpr.CustomObject) error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("create").Inc()

	clusterCustomer := resources.ClusterCustomer(customObject)
	clusterID := resources.ClusterID(customObject)
	clusterNamespace := resources.ClusterNamespace(customObject)

	s.logger.Log("debug", fmt.Sprintf("creating namespace '%s' for cluster '%s'", clusterNamespace, clusterID))

	namespace := v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: clusterNamespace,
			Labels: map[string]string{
				"cluster":  clusterID,
				"customer": clusterCustomer,
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

func (s *Service) deleteNamespace(customObject kvmtpr.CustomObject) error {
	start := time.Now()
	namespaceActionTotal.WithLabelValues("delete").Inc()

	clusterID := resources.ClusterID(customObject)
	clusterNamespace := resources.ClusterNamespace(customObject)

	s.logger.Log("debug", fmt.Sprintf("deleting namespace '%s' for cluster '%s'", clusterNamespace, clusterID))

	err := s.kubernetesClient.Core().Namespaces().Delete(clusterNamespace, v1.NewDeleteOptions(0))
	if errors.IsNotFound(err) {
		s.logger.Log("debug", "namespace does not exist")
	} else if err != nil {
		return err
	}

	s.logger.Log("debug", "namespace deleted")
	namespaceActionTime.WithLabelValues("delete").Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
