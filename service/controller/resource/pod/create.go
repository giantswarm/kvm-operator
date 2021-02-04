package pod

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	pod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "determining readiness for node pod")

	var nodeIP string
	var serviceName string
	{
		var ok bool
		nodeIP, ok = pod.Annotations[key.AnnotationIp]
		if !ok || nodeIP == "" {
			r.logger.Debugf(ctx, "node pod has no ip annotation %#q, skipping", key.AnnotationIp)
			return nil
		}

		serviceName, ok = pod.Annotations[key.AnnotationService]
		if !ok || serviceName == "" {
			r.logger.Debugf(ctx, "node pod has no service annotation %#q, skipping", key.AnnotationService)
			return nil
		} else if serviceName == key.MasterID {
			r.logger.Debugf(ctx, "node pod contains a workload cluster master node, skipping")
			return nil
		}
	}

	var readyForTraffic bool
	{
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				readyForTraffic = condition.Status == corev1.ConditionTrue
				break
			}
		}

		statusAnnotation, hasStatusAnnotation := pod.Annotations[key.AnnotationPodNodeStatus]
		if hasStatusAnnotation && statusAnnotation != "" {
			readyForTraffic = readyForTraffic && statusAnnotation == key.PodNodeStatusReady
		}
	}

	if readyForTraffic {
		r.logger.Debugf(ctx, "determined node pod is ready")
	} else {
		r.logger.Debugf(ctx, "determined node pod is not ready")
	}

	var endpoints *corev1.Endpoints
	{
		r.logger.Debugf(ctx, "retrieving endpoints %#q", serviceName)
		var err error
		endpoints, err = r.k8sClient.CoreV1().Endpoints(pod.Namespace).Get(ctx, serviceName, v1.GetOptions{})
		if err != nil {
			r.logger.Debugf(ctx, "error retrieving endpoints")
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "retrieved endpoints")
	}

	var updated bool
	addresses := endpoints.Subsets[0].Addresses
	if readyForTraffic {
		addresses, updated = addToEndpoints(addresses, nodeIP)
		if updated {
			r.logger.Debugf(ctx, "node ip %#q is not in endpoints but pod and node are ready, adding ip", nodeIP)
		}
	} else {
		addresses, updated = removeFromEndpoints(addresses, nodeIP)
		if updated {
			r.logger.Debugf(ctx, "node ip %#q is in endpoints but pod or node is not ready, removing ip", nodeIP)
		}
	}

	if updated {
		r.logger.Debugf(ctx, "updating endpoints", pod.Namespace, pod.Name)
		endpoints.Subsets[0].Addresses = addresses
		_, err := r.k8sClient.CoreV1().Endpoints(pod.Namespace).Update(ctx, endpoints, v1.UpdateOptions{})
		if err != nil {
			r.logger.Debugf(ctx, "error updating endpoints", pod.Namespace, pod.Name)
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "updated endpoints", pod.Namespace, pod.Name)
	} else {
		r.logger.Debugf(ctx, "not updating endpoints")
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
