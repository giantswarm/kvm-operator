package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	pod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "determining readiness for node pod")

	var nodeIP string
	var serviceName types.NamespacedName
	{
		var ok bool
		nodeIP, ok = pod.Annotations[key.AnnotationIp]
		if !ok || nodeIP == "" {
			r.logger.Debugf(ctx, "node pod has no ip annotation %#q, skipping", key.AnnotationIp)
			return nil
		}

		serviceNameAnnotation, ok := pod.Annotations[key.AnnotationService]
		if !ok || serviceNameAnnotation == "" {
			r.logger.Debugf(ctx, "node pod has no service annotation %#q, skipping", key.AnnotationService)
			return nil
		} else if serviceNameAnnotation == key.MasterID {
			r.logger.Debugf(ctx, "node pod contains a workload cluster master node, skipping")
			return nil
		}

		serviceName = types.NamespacedName{
			Namespace: pod.Namespace,
			Name:      serviceNameAnnotation,
		}
	}

	readyForTraffic := key.PodIsReady(pod)
	if annotation := pod.Annotations[key.AnnotationPodNodeStatus]; annotation != "" {
		readyForTraffic = readyForTraffic && annotation == key.PodNodeStatusReady
	}

	if readyForTraffic {
		r.logger.Debugf(ctx, "determined node pod is ready")
	} else {
		r.logger.Debugf(ctx, "determined node pod is not ready")
	}

	var needCreate bool
	var endpoints *corev1.Endpoints
	{
		r.logger.Debugf(ctx, "retrieving endpoints %#q", serviceName)
		var err error
		endpoints, err = r.k8sClient.CoreV1().Endpoints(serviceName.Namespace).Get(ctx, serviceName.Name, v1.GetOptions{})
		if errors.IsNotFound(err) {
			needCreate = true
		} else {
			r.logger.Debugf(ctx, "error retrieving endpoints")
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "retrieved endpoints")
	}

	if needCreate {
		err := r.createEndpoints(ctx, serviceName, readyForTraffic, nodeIP)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		err := r.updateEndpoints(ctx, endpoints, readyForTraffic, nodeIP)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) createEndpoints(ctx context.Context, serviceName types.NamespacedName, readyForTraffic bool, nodeIP string) error {
	service, err := r.k8sClient.CoreV1().Services(serviceName.Namespace).Get(ctx, serviceName.Name, v1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	var endpoints corev1.Endpoints
	{
		var ports []corev1.EndpointPort
		for _, s := range service.Spec.Ports {
			ports = append(ports, corev1.EndpointPort{
				Name:        s.Name,
				Port:        s.Port,
				Protocol:    s.Protocol,
				AppProtocol: s.AppProtocol,
			})
		}

		var addresses []corev1.EndpointAddress
		var notReadyAddresses []corev1.EndpointAddress
		address := corev1.EndpointAddress{IP: nodeIP}
		if readyForTraffic {
			addresses = append(addresses, address)
		} else {
			notReadyAddresses = append(notReadyAddresses, address)
		}

		endpoints = corev1.Endpoints{
			ObjectMeta: v1.ObjectMeta{
				Name:      serviceName.Name,
				Namespace: serviceName.Namespace,
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses:         addresses,
					NotReadyAddresses: notReadyAddresses,
					Ports:             ports,
				},
			},
		}
	}

	_, err = r.k8sClient.CoreV1().Endpoints(serviceName.Namespace).Create(ctx, &endpoints, v1.CreateOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) updateEndpoints(ctx context.Context, endpoints *corev1.Endpoints, readyForTraffic bool, nodeIP string) error {
	var updated bool
	addresses := endpoints.Subsets[0].Addresses
	notReadyAddresses := endpoints.Subsets[0].NotReadyAddresses
	if readyForTraffic {
		addresses, updated = addToAddresses(addresses, nodeIP)
		notReadyAddresses, _ = removeFromAddresses(notReadyAddresses, nodeIP)
		if updated {
			r.logger.Debugf(ctx, "node ip %#q is not in endpoints but pod and node are ready, adding ip", nodeIP)
		}
	} else {
		addresses, updated = removeFromAddresses(addresses, nodeIP)
		notReadyAddresses, _ = addToAddresses(notReadyAddresses, nodeIP)
		if updated {
			r.logger.Debugf(ctx, "node ip %#q is in endpoints but pod or node is not ready, removing ip", nodeIP)
		}
	}

	if updated {
		r.logger.Debugf(ctx, "updating endpoints")
		endpoints.Subsets[0].Addresses = addresses
		endpoints.Subsets[0].NotReadyAddresses = notReadyAddresses
		_, err := r.k8sClient.CoreV1().Endpoints(endpoints.Namespace).Update(ctx, endpoints, v1.UpdateOptions{})
		if err != nil {
			r.logger.Debugf(ctx, "error updating endpoints")
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "updated endpoints")
	} else {
		r.logger.Debugf(ctx, "not updating endpoints")
	}

	return nil
}

func addToAddresses(addresses []corev1.EndpointAddress, nodeIP string) ([]corev1.EndpointAddress, bool) {
	for _, address := range addresses {
		if address.IP == nodeIP {
			return addresses, false
		}
	}
	return append(addresses, corev1.EndpointAddress{IP: nodeIP}), true
}

func removeFromAddresses(addresses []corev1.EndpointAddress, nodeIP string) ([]corev1.EndpointAddress, bool) {
	for i, address := range addresses {
		if address.IP == nodeIP {
			return append(addresses[:i], addresses[i+1:]...), true
		}
	}
	return addresses, false
}
