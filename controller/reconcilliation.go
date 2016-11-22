package controller

import (
	golang_errors "errors"
	"log"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

var (
	unknownResourceTypeErr = golang_errors.New("Unknown resource type")

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

type action interface {
	Run(kubernetes.Interface) error
}

type create struct {
	resource runtime.Object
}

func (a *create) Run(clientset kubernetes.Interface) error {
	switch res := a.resource.(type) {
	case *v1.Service:
		log.Printf("creating service %v/%v\n", res.Namespace, res.Name)
		_, err := clientset.Core().Services(res.Namespace).Create(res)
		return err
	default:
		return unknownResourceTypeErr
	}
}

type update struct {
	resource runtime.Object
}

func (a *update) Run(clientset kubernetes.Interface) error {
	switch res := a.resource.(type) {
	case *v1.Service:
		log.Printf("updating service %v/%v\n", res.Namespace, res.Name)
		_, err := clientset.Core().Services(res.Namespace).Update(res)
		return err
	default:
		return unknownResourceTypeErr
	}
}

func init() {
	prometheus.MustRegister(reconcilliationTotal)
	prometheus.MustRegister(reconicilliationTime)
	prometheus.MustRegister(reconcilliationResourceModificationTotal)
	prometheus.MustRegister(reconcilliationResourceModificationTime)
}

func (c *controller) getExistingResource(resource runtime.Object) (runtime.Object, error) {
	switch res := resource.(type) {
	case *v1.Service:
		return c.clientset.Core().Services(res.Namespace).Get(res.Name)
	default:
		return nil, unknownResourceTypeErr
	}
}

// reconcileResourceState takes a list of Kubernetes resources, and makes sure
// that these resources exist.
func (c *controller) reconcileResourceState(namespaceName string, resources []runtime.Object) error {
	start := time.Now()
	reconcilliationTotal.WithLabelValues(namespaceName).Inc()

	log.Println("starting reconcilliation for namespace:", namespaceName)

	actions := []action{}

	for _, resource := range resources {
		existingResource, err := c.getExistingResource(resource)

		if errors.IsNotFound(err) {
			createAction := &create{resource: resource}
			actions = append(actions, createAction)
			continue
		}

		if err != nil {
			return err
		}

		if !reflect.DeepEqual(existingResource, resource) {
			updateAction := &update{
				resource: resource,
			}
			actions = append(actions, updateAction)
			continue
		}
	}

	for _, action := range actions {
		if err := action.Run(c.clientset); err != nil {
			return err
		}
	}

	log.Println("finished reconcilliation for namespace:", namespaceName)

	reconicilliationTime.WithLabelValues(namespaceName).Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
