package endpoint

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	IPAnnotation      = "endpoint.kvm.giantswarm.io/ip"
	Name              = "endpointv24"
	ServiceAnnotation = "endpoint.kvm.giantswarm.io/service"
)

type Config struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) newK8sEndpoint(endpoint *Endpoint) *corev1.Endpoints {
	k8sEndpoint := &corev1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      endpoint.ServiceName,
			Namespace: endpoint.ServiceNamespace,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: endpoint.Addresses,
				Ports:     endpoint.Ports,
			},
		},
	}

	return k8sEndpoint
}

func containsIP(ips []string, ip string) bool {
	for _, foundIP := range ips {
		if foundIP == ip {
			return true
		}
	}

	return false
}

func getAnnotations(pod corev1.Pod, ipAnnotationName string, serviceAnnotationName string) (ipAnnotationValue string, serviceAnnotationValue string, err error) {
	ipAnnotationValue, ok := pod.GetAnnotations()[ipAnnotationName]
	if !ok {
		return "", "", microerror.Maskf(missingAnnotationError, "expected annotation '%s' to be set", ipAnnotationName)
	}
	if ipAnnotationValue == "" {
		return "", "", microerror.Maskf(missingAnnotationError, "empty annotation '%s'", ipAnnotationName)
	}
	serviceAnnotationValue, ok = pod.GetAnnotations()[serviceAnnotationName]
	if !ok {
		return "", "", microerror.Maskf(missingAnnotationError, "expected annotation '%s' to be set", serviceAnnotationName)
	}
	return ipAnnotationValue, serviceAnnotationValue, nil
}

func ipsToAddresses(ips []string) []corev1.EndpointAddress {
	var addresses []corev1.EndpointAddress

	for _, ip := range ips {
		k8sAddress := corev1.EndpointAddress{
			IP: ip,
		}
		addresses = append(addresses, k8sAddress)
	}

	return addresses
}

func isEmptyEndpoint(endpoint *corev1.Endpoints) bool {
	if endpoint == nil {
		return true
	}

	for _, subset := range endpoint.Subsets {
		if len(subset.Addresses) > 0 {
			return false
		}
	}

	return true
}

func removeIP(ips []string, ip string) []string {
	for index, foundIP := range ips {
		if foundIP == ip {
			return append(ips[:index], ips[index+1:]...)
		}
	}
	return ips
}

func toEndpoint(v interface{}) (*Endpoint, error) {
	if v == nil {
		return nil, nil
	}
	endpoint, ok := v.(*Endpoint)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &Endpoint{}, v)
	}
	return endpoint, nil
}

func toK8sEndpoint(v interface{}) (*corev1.Endpoints, error) {
	if v == nil {
		return nil, nil
	}
	k8sEndpoint, ok := v.(*corev1.Endpoints)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.Endpoints{}, v)
	}
	return k8sEndpoint, nil
}

func toPod(v interface{}) (*corev1.Pod, error) {
	if v == nil {
		return nil, nil
	}

	pod, ok := v.(*corev1.Pod)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.Pod{}, v)
	}

	return pod, nil
}
