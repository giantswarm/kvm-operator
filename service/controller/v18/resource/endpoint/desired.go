package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	podIsReady := false
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			podIsReady = true
			break
		}
	}
	// If the pod is not ready so we should not add the ip to the endpoint list.
	if !podIsReady {
		return nil, nil
	}
	// If pod has deletionTimestamp consider it dead and remove endpoint ip.
	if pod.DeletionTimestamp != nil {
		return nil, nil
	}

	endpointIP, serviceName, err := getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint := &Endpoint{
		IPs: []string{
			endpointIP,
		},
		ServiceName:      serviceName,
		ServiceNamespace: pod.GetNamespace(),
	}

	return desiredEndpoint, nil
}
