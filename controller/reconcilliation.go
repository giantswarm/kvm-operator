package controller

import (
	golang_errors "errors"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/serializer/json"
	"k8s.io/client-go/pkg/util/strategicpatch"
)

const (
	ClusterControllerConfigurationAnnotation = "cluster-controller.giantswarm.io/last-configuration"
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

func (c *controller) addConfigurationAnnotation(resource runtime.Object) (runtime.Object, error) {
	serializer := json.NewSerializer(json.DefaultMetaFactory, api.Scheme, api.Scheme, false)

	encoded, err := runtime.Encode(serializer, resource)
	if err != nil {
		return nil, err
	}

	switch res := resource.(type) {
	case *v1.ConfigMap:
		res.Annotations = map[string]string{}
		res.Annotations[ClusterControllerConfigurationAnnotation] = string(encoded)
		return res, nil
	case *v1.Service:
		res.Annotations = map[string]string{}
		res.Annotations[ClusterControllerConfigurationAnnotation] = string(encoded)
		return res, nil
	case *v1beta1.Deployment:
		res.Annotations = map[string]string{}
		res.Annotations[ClusterControllerConfigurationAnnotation] = string(encoded)
		return res, nil
	case *v1beta1.Ingress:
		res.Annotations = map[string]string{}
		res.Annotations[ClusterControllerConfigurationAnnotation] = string(encoded)
		return res, nil
	case *v1beta1.Job:
		res.Annotations = map[string]string{}
		res.Annotations[ClusterControllerConfigurationAnnotation] = string(encoded)
		return res, nil
	default:
		return nil, unknownResourceTypeErr
	}
}

func (c *controller) getConfigurationAnnotation(resource runtime.Object) (string, bool) {
	log.Println("getting config annotation")
	log.Println("getConfigurationAnnotation: resource:", resource)

	switch res := resource.(type) {
	case *v1.ConfigMap:
		s, ok := res.Annotations[ClusterControllerConfigurationAnnotation]
		return s, ok
	case *v1.Service:
		log.Println("annotations:", res.Annotations)
		s, ok := res.Annotations[ClusterControllerConfigurationAnnotation]
		return s, ok
	case *v1beta1.Deployment:
		s, ok := res.Annotations[ClusterControllerConfigurationAnnotation]
		return s, ok
	case *v1beta1.Ingress:
		s, ok := res.Annotations[ClusterControllerConfigurationAnnotation]
		return s, ok
	case *v1beta1.Job:
		s, ok := res.Annotations[ClusterControllerConfigurationAnnotation]
		return s, ok
	default:
		return "", false
	}
}

func (c *controller) createResource(resource runtime.Object) error {
	var err error = nil

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

func (c *controller) resourceNeedsUpdate(old runtime.Object, new runtime.Object) bool {
	oldConfiguration, _ := c.getConfigurationAnnotation(old)
	newConfiguration, _ := c.getConfigurationAnnotation(new)

	return !(oldConfiguration == newConfiguration)
}

func (c *controller) updateResource(old runtime.Object, new runtime.Object) error {
	log.Println("update resource: old:", old)

	serializer := json.NewSerializer(json.DefaultMetaFactory, api.Scheme, api.Scheme, false)

	// Encode the old and new resources
	oldEncoded, err := runtime.Encode(serializer, old)
	if err != nil {
		return err
	}
	newEncoded, err := runtime.Encode(serializer, new)
	if err != nil {
		return err
	}

	configurationAnnotation, ok := c.getConfigurationAnnotation(old)
	if !ok {
		return golang_errors.New("Could not get annotation from resource")
	}

	patch, err := strategicpatch.CreateThreeWayMergePatch([]byte(configurationAnnotation), newEncoded, oldEncoded, old, false)
	if err != nil {
		return err
	}
	patchedResource, err := strategicpatch.StrategicMergePatch(oldEncoded, patch, old)
	if err != nil {
		return err
	}

	// Decode the resource into an Object
	newObject, err := runtime.Decode(serializer, patchedResource)
	if err != nil {
		return err
	}

	// Update the object
	var kerr error = nil
	switch res := newObject.(type) {
	case *v1.ConfigMap:
		_, kerr = c.clientset.Core().ConfigMaps(res.Namespace).Update(res)
	case *v1.Service:
		_, kerr = c.clientset.Core().Services(res.Namespace).Update(res)
	case *v1beta1.Deployment:
		_, kerr = c.clientset.Extensions().Deployments(res.Namespace).Update(res)
	case *v1beta1.Ingress:
		_, kerr = c.clientset.Extensions().Ingresses(res.Namespace).Update(res)
	case *v1beta1.Job:
		_, kerr = c.clientset.Extensions().Jobs(res.Namespace).Update(res)
	default:
		kerr = unknownResourceTypeErr
	}

	return kerr
}

// reconcileResourceState takes a list of resources, which represent the desired
// state of the underlying cluster resources, and performs the necessary
// actions to bring the actual state of the cluster resources closer to this
// desired state.
func (c *controller) reconcileResourceState(namespaceName string, resources []runtime.Object) error {
	start := time.Now()
	reconcilliationTotal.WithLabelValues(namespaceName).Inc()

	log.Println("starting reconcilliation for namespace:", namespaceName)

	for _, resource := range resources {
		existingResource, err := c.getExistingResource(resource)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		kind, namespace, name, infoErr := c.getInfoForResource(resource)
		if infoErr != nil {
			return infoErr
		}

		resource, err := c.addConfigurationAnnotation(resource)
		if err != nil {
			return err
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

		if existingResource != nil && c.resourceNeedsUpdate(existingResource, resource) {
			start := time.Now()
			reconcilliationResourceModificationTotal.WithLabelValues(namespace, kind, "updated").Inc()

			log.Printf("updating %v %v/%v\n", kind, namespace, name)
			if err := c.updateResource(existingResource, resource); err != nil {
				return err
			}

			reconcilliationResourceModificationTime.WithLabelValues(namespace, kind, "updated").Set(float64(time.Since(start) / time.Millisecond))

			continue
		}
	}

	log.Println("finished reconcilliation for namespace:", namespaceName)

	reconicilliationTime.WithLabelValues(namespaceName).Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
