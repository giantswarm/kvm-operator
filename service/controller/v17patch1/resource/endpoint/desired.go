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

	endpointIP, serviceName, err := getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint := &Endpoint{
		IPs: []string{
			endpointIP,
		},
		RemoveEndpoint:   false,
		ServiceName:      serviceName,
		ServiceNamespace: pod.GetNamespace(),
	}

	podIsReady := false
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			podIsReady = true
			break
		}
	}
	// If pod has deletionTimestamp consider it dead and remove endpoint ip.
	if pod.DeletionTimestamp != nil {
		desiredEndpoint.RemoveEndpoint = true
		r.logger.Log("deletion timestamp")
	}

	// If the pod is not ready so we should not add the ip to the endpoint list.
	if !podIsReady && pod.DeletionTimestamp == nil {
		desiredEndpoint.IPs = []string{}
	}

	return desiredEndpoint, nil
}
