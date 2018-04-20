package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(createState)
	if err != nil {
		return microerror.Mask(err)
	}
	if k8sEndpoint == nil {
		return nil // Nothing to do.
	}

	_, err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if currentEndpoint != nil {
		return nil, nil // An endpoint exists already, we should update instead of create.
	}

	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if desiredEndpoint == nil {
		return nil, nil // Nothing to do.
	}

	endpoint := &Endpoint{
		ServiceName:      desiredEndpoint.ServiceName,
		ServiceNamespace: desiredEndpoint.ServiceNamespace,
	}
	for _, desiredIP := range desiredEndpoint.IPs {
		if !containsIP(endpoint.IPs, desiredIP) {
			endpoint.IPs = append(endpoint.IPs, desiredIP)
		}
	}

	if len(endpoint.IPs) == 0 {
		return nil, nil // Nothing to do.
	}

	createState, err := r.newK8sEndpoint(endpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return createState, nil
}
