package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if k8sEndpoint == nil {
		return nil // Nothing to do.
	}
	if isEmptyEndpoint(*k8sEndpoint) {
		return nil
	}

	_, err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Update(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	createState, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	updateState, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()

	patch.SetCreateChange(createState)
	patch.SetUpdateChange(updateState)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if currentEndpoint == nil {
		return nil, nil // The endpoint does not exist, we should create it instead.
	}

	var updateChange *corev1.Endpoints
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the endpoint has to be computed")

		var ips []string
		for _, currentIP := range currentEndpoint.IPs {
			if !containsIP(ips, currentIP) {
				ips = append(ips, currentIP)
			}
		}
		for _, desiredIP := range desiredEndpoint.IPs {
			if !containsIP(ips, desiredIP) {
				ips = append(ips, desiredIP)
			}
		}

		if len(ips) != 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the endpoint has to be computed")

			endpoint := &Endpoint{
				ServiceName:      desiredEndpoint.ServiceName,
				ServiceNamespace: desiredEndpoint.ServiceNamespace,
				IPs:              ips,
			}
			updateChange, err = r.newK8sEndpoint(endpoint)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the endpoint does not have to be computed")
		}
	}

	return updateChange, nil
}
