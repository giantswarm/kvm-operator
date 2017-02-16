package operator

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

var (
	reconcilliationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reconcilliation_total",
			Help: "Number of times we have performed reconcilliation for a namespace",
		},
		[]string{"namespace"},
	)
	reconicilliationTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "reconcilliation_milliseconds",
			Help: "Time taken to handle resource reconcilliation for a namespace, in milliseconds",
		},
		[]string{"namespace"},
	)

	reconcilliationResourceModificationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reconcilliation_resource_modification_total",
			Help: "Number of times a resource has been modified during reconcilliation",
		},
		[]string{"namespace", "kind", "action"},
	)
	reconcilliationResourceModificationTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "reconcilliation_resource_modification_time",
			Help: "Time taken to handle resource modification for reconcilliation, in milliseconds",
		},
		[]string{"namespace", "kind", "action"},
	)
)

func init() {
	prometheus.MustRegister(reconcilliationTotal)
	prometheus.MustRegister(reconicilliationTime)
	prometheus.MustRegister(reconcilliationResourceModificationTotal)
	prometheus.MustRegister(reconcilliationResourceModificationTime)
}

// reconcileResourceState takes a list of Kubernetes resources, and makes sure
// that these resources exist.
func (s *Service) reconcileResourceState(namespaceName string, resources []runtime.Object) error {
	start := time.Now()
	reconcilliationTotal.WithLabelValues(namespaceName).Inc()

	s.logger.Log("debug", fmt.Sprintf("starting reconcilliation for namespace '%s'", namespaceName))

	for _, resource := range resources {
		switch t := resource.(type) {
		case *v1.ConfigMap:
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespaceName, "configmap", "created").Inc()
			s.logger.Log("debug", fmt.Sprintf("creating configmap '%s'", t.Name))
			if _, err := s.kubernetesClient.Core().ConfigMaps(namespaceName).Create(t); err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
			reconcilliationResourceModificationTime.WithLabelValues(namespaceName, "configmap", "created").Set(float64(time.Since(start) / time.Millisecond))

		case *v1.Service:
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespaceName, "service", "created").Inc()
			s.logger.Log("debug", fmt.Sprintf("creating service '%s'", t.Name))
			if _, err := s.kubernetesClient.Core().Services(namespaceName).Create(t); err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
			reconcilliationResourceModificationTime.WithLabelValues(namespaceName, "service", "created").Set(float64(time.Since(start) / time.Millisecond))

		case *v1beta1.Deployment:
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespaceName, "deployment", "created").Inc()
			s.logger.Log("debug", fmt.Sprintf("creating deployment '%s'", t.Name))
			if _, err := s.kubernetesClient.Extensions().Deployments(namespaceName).Create(t); err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
			reconcilliationResourceModificationTime.WithLabelValues(namespaceName, "deployment", "created").Set(float64(time.Since(start) / time.Millisecond))

		case *v1beta1.Ingress:
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespaceName, "ingress", "created").Inc()
			s.logger.Log("debug", fmt.Sprintf("creating ingress '%s'", t.Name))
			if _, err := s.kubernetesClient.Extensions().Ingresses(namespaceName).Create(t); err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
			reconcilliationResourceModificationTime.WithLabelValues(namespaceName, "ingress", "created").Set(float64(time.Since(start) / time.Millisecond))

		case *v1beta1.Job:
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespaceName, "job", "created").Inc()
			s.logger.Log("debug", fmt.Sprintf("creating job '%s'", t.Name))
			if _, err := s.kubernetesClient.Extensions().Jobs(namespaceName).Create(t); err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
			reconcilliationResourceModificationTime.WithLabelValues(namespaceName, "job", "created").Set(float64(time.Since(start) / time.Millisecond))

		default:
			s.logger.Log("error", fmt.Sprintf("unknown type '%T'", t))
		}
	}

	s.logger.Log("debug", fmt.Sprintf("finished reconcilliation for namespace '%s'", namespaceName))

	reconicilliationTime.WithLabelValues(namespaceName).Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
