package controller

import (
	golang_errors "errors"
	"log"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
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

func init() {
	prometheus.MustRegister(reconcilliationTotal)
	prometheus.MustRegister(reconicilliationTime)
	prometheus.MustRegister(reconcilliationResourceModificationTotal)
	prometheus.MustRegister(reconcilliationResourceModificationTime)
}

func (c *controller) getInfoForResource(resource runtime.Object) (string, string, string, error) {
	switch res := resource.(type) {
	case *v1.ConfigMap:
		return "configmap", res.Namespace, res.Name, nil
	case *v1.Service:
		return "service", res.Namespace, res.Name, nil
	case *v1beta1.Deployment:
		return "deployment", res.Namespace, res.Name, nil
	case *v1beta1.Ingress:
		return "ingress", res.Namespace, res.Name, nil
	case *v1beta1.Job:
		return "job", res.Namespace, res.Name, nil
	default:
		return "", "", "", unknownResourceTypeErr
	}
}

// getExistingResource takes a resource, and returns the resource that Kubernetes has.
func (c *controller) getExistingResource(resource runtime.Object) (runtime.Object, error) {
	switch res := resource.(type) {
	case *v1.ConfigMap:
		return c.clientset.Core().ConfigMaps(res.Namespace).Get(res.Name)
	case *v1.Service:
		return c.clientset.Core().Services(res.Namespace).Get(res.Name)
	case *v1beta1.Deployment:
		return c.clientset.Extensions().Deployments(res.Namespace).Get(res.Name)
	case *v1beta1.Ingress:
		return c.clientset.Extensions().Ingresses(res.Namespace).Get(res.Name)
	case *v1beta1.Job:
		return c.clientset.Extensions().Jobs(res.Namespace).Get(res.Name)
	default:
		return nil, unknownResourceTypeErr
	}
}

func (c *controller) createResource(resource runtime.Object) error {
	var err error

	switch res := resource.(type) {
	case *v1.ConfigMap:
		_, err = c.clientset.Core().ConfigMaps(res.Namespace).Create(res)
	case *v1.Service:
		_, err = c.clientset.Core().Services(res.Namespace).Create(res)
	case *v1beta1.Deployment:
		_, err = c.clientset.Extensions().Deployments(res.Namespace).Create(res)
	case *v1beta1.Ingress:
		_, err = c.clientset.Extensions().Ingresses(res.Namespace).Create(res)
	case *v1beta1.Job:
		_, err = c.clientset.Extensions().Jobs(res.Namespace).Create(res)
	default:
		err = unknownResourceTypeErr
	}

	return err
}

func (c *controller) updateResource(resource runtime.Object) error {
	var err error

	switch res := resource.(type) {
	case *v1.ConfigMap:
		_, err = c.clientset.Core().ConfigMaps(res.Namespace).Update(res)
	case *v1.Service:
		_, err = c.clientset.Core().Services(res.Namespace).Update(res)
	case *v1beta1.Deployment:
		_, err = c.clientset.Extensions().Deployments(res.Namespace).Update(res)
	case *v1beta1.Ingress:
		_, err = c.clientset.Extensions().Ingresses(res.Namespace).Update(res)
	case *v1beta1.Job:
		_, err = c.clientset.Extensions().Jobs(res.Namespace).Update(res)
	default:
		err = unknownResourceTypeErr
	}

	return err
}

// reconcileResourceState takes a list of resources, which represent the desired
// state of the underlying cluster resources, and performs the necessary
// actions to bring the actual state of the cluster resources closer to this
// desired state.
func (c *controller) reconcileResourceState(namespaceName string, resources []runtime.Object) error {
	start := time.Now()
	reconcilliationTotal.WithLabelValues(namespaceName).Inc()

	log.Println("starting reconcilliation for:", namespaceName)

	for _, resource := range resources {
		existingResource, err := c.getExistingResource(resource)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		kind, namespace, name, infoErr := c.getInfoForResource(resource)
		if infoErr != nil {
			return infoErr
		}

		if errors.IsNotFound(err) {
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespace, kind, "created").Inc()

			log.Printf("creating %v %v/%v\n", kind, namespace, name)
			if err := c.createResource(resource); err != nil {
				return err
			}

			reconcilliationResourceModificationTime.WithLabelValues(namespace, kind, "created").Set(float64(time.Since(start) / time.Millisecond))

			continue
		}

		// If the actual state of the resource does not match the desired state, update it
		if !reflect.DeepEqual(existingResource, resource) {
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespace, kind, "updated").Inc()

			log.Printf("updating %v %v/%v\n", kind, namespace, name)
			if err := c.updateResource(resource); err != nil {
				return err
			}

			reconcilliationResourceModificationTime.WithLabelValues(namespace, kind, "updated").Set(float64(time.Since(start) / time.Millisecond))

			continue
		}
	}

	log.Println("finished reconcilliation for:", namespaceName)

	reconicilliationTime.WithLabelValues(namespaceName).Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
