package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	pod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "determining readiness for node pod")

	nodeIP, serviceName, err := r.podEndpointData(ctx, pod)
	if IsMissingAnnotation(err) {
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var readyForTraffic bool
	if serviceName.Name == key.MasterID {
		readyForTraffic = true
	} else {
		readyForTraffic = key.PodIsReady(pod) && key.PodNodeIsReady(pod)
	}

	if readyForTraffic {
		r.logger.Debugf(ctx, "determined node pod is ready")
	} else {
		r.logger.Debugf(ctx, "determined node pod is not ready")
	}

	var service corev1.Service
	{
		err := r.ctrlClient.Get(ctx, serviceName, &service)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var needCreate bool
	var endpoints corev1.Endpoints
	{
		r.logger.Debugf(ctx, "retrieving endpoints %#q", serviceName)
		err = r.ctrlClient.Get(ctx, serviceName, &endpoints)
		if errors.IsNotFound(err) {
			r.logger.Debugf(ctx, "endpoints not found, creating")
			needCreate = true
		} else if err != nil {
			r.logger.Debugf(ctx, "error retrieving endpoints")
			return microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "retrieved endpoints")
		}
	}

	if needCreate {
		err := r.createEndpoints(ctx, service, readyForTraffic, nodeIP)
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

func (r *Resource) createEndpoints(ctx context.Context, service corev1.Service, readyForTraffic bool, nodeIP string) error {
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
				Name:      service.Name,
				Namespace: service.Namespace,
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

	err := r.ctrlClient.Create(ctx, &endpoints)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) updateEndpoints(ctx context.Context, endpoints corev1.Endpoints, readyForTraffic bool, nodeIP string) error {
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
		err := r.ctrlClient.Update(ctx, &endpoints)
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
