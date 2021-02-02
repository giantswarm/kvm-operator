package pod

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return microerror.Mask(invalidConfigError)
	}

	nodeIP, ok := pod.Annotations["endpoint.kvm.giantswarm.io/ip"]
	if !ok || nodeIP == "" {
		return nil
	}

	statusAnnotation, ok := pod.Annotations["kvm-operator.giantswarm.io/node-status"]
	if !ok || statusAnnotation == "" {
		return nil
	}

	endpointsName, ok := pod.Annotations["endpoint.kvm.giantswarm.io/service"]
	if !ok || endpointsName == "" {
		return nil
	}

	endpoints, err := r.k8sClient.CoreV1().Endpoints(pod.Namespace).Get(ctx, endpointsName, v1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	var updated bool
	addresses := endpoints.Subsets[0].Addresses
	if statusAnnotation == "not-ready" {
		addresses, updated = removeFromEndpoints(addresses, nodeIP)
	} else if statusAnnotation == "ready" {
		addresses, updated = addToEndpoints(addresses, nodeIP)
	}

	if updated {
		endpoints.Subsets[0].Addresses = addresses
		_, err := r.k8sClient.CoreV1().Endpoints(pod.Namespace).Update(ctx, endpoints, v1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func addToEndpoints(addresses []corev1.EndpointAddress, nodeIP string) ([]corev1.EndpointAddress, bool) {
	for _, address := range addresses {
		if address.IP == nodeIP {
			return addresses, false
		}
	}
	return append(addresses, corev1.EndpointAddress{IP: nodeIP}), true
}

func removeFromEndpoints(addresses []corev1.EndpointAddress, nodeIP string) ([]corev1.EndpointAddress, bool) {
	for i, address := range addresses {
		if address.IP == nodeIP {
			return append(addresses[:i], addresses[i+1:]...), true
		}
	}
	return addresses, false
}
