package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if k8sEndpoint != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the endpoint in the Kubernetes API")

		_, err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(k8sEndpoint)
		if errors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the endpoint in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the endpoint does not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createChange *corev1.Endpoints
	{
		var ips []string
		for _, desiredIP := range desiredEndpoint.IPs {
			if !containsIP(ips, desiredIP) {
				ips = append(ips, desiredIP)
			}
		}

		endpointExists := currentEndpoint != nil
		ipsEmpty := len(ips) == 0

		if !endpointExists && !ipsEmpty {
			endpoint := &Endpoint{
				ServiceName:      desiredEndpoint.ServiceName,
				ServiceNamespace: desiredEndpoint.ServiceNamespace,
				IPs:              ips,
			}
			createChange, err = r.newK8sEndpoint(endpoint)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return createChange, nil
}
