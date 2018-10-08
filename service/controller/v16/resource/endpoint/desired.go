package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("level", "info", "message", "endpoint get Desired State loop", "extra", pod.Name)

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
