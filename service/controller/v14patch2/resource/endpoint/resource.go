package endpoint

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	IPAnnotation      = "endpoint.kvm.giantswarm.io/ip"
	Name              = "endpointv14patch1"
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

func (r *Resource) newK8sEndpoint(endpoint *Endpoint) (*corev1.Endpoints, error) {
	k8sAddresses := []corev1.EndpointAddress{}
	for _, endpointIP := range endpoint.IPs {
		k8sAddress := corev1.EndpointAddress{
			IP: endpointIP,
		}
		k8sAddresses = append(k8sAddresses, k8sAddress)
	}

	k8sService, err := r.k8sClient.CoreV1().Services(endpoint.ServiceNamespace).Get(endpoint.ServiceName, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

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
				Ports: serviceToPorts(k8sService),
			},
		},
	}

	for i := range k8sEndpoint.Subsets {
		k8sEndpoint.Subsets[i].Addresses = k8sAddresses
	}

	return k8sEndpoint, nil
}

func cutIPs(base []string, cutset []string) []string {
	resultIPs := []string{}
	// Deduplicate entries from base.
	for _, baseIP := range base {
		if !containsIP(resultIPs, baseIP) {
			resultIPs = append(resultIPs, baseIP)
		}
	}
	// Cut the cutset out of base.
	for _, cutsetIP := range cutset {
		if containsIP(resultIPs, cutsetIP) {
			resultIPs = removeIP(resultIPs, cutsetIP)
		}
	}
	return resultIPs
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

func isEmptyEndpoint(endpoint corev1.Endpoints) bool {
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

func serviceToPorts(s *corev1.Service) []corev1.EndpointPort {
	var ports []corev1.EndpointPort

	for _, p := range s.Spec.Ports {
		port := corev1.EndpointPort{
			Name: p.Name,
			Port: p.Port,
		}

		ports = append(ports, port)
	}

	return ports
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
