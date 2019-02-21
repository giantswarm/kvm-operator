package endpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	endpointToUpdate, err := toK8sEndpoint(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if !isEmptyEndpoint(endpointToUpdate) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating endpoint '%s'", endpointToUpdate.GetName()))

		_, err = r.k8sClient.CoreV1().Endpoints(endpointToUpdate.Namespace).Update(endpointToUpdate)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated endpoint '%s'", endpointToUpdate.GetName()))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not updating endpoint '%s'", endpointToUpdate.GetName()))
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

	var updateChange *corev1.Endpoints
	{
		ips := ipsForUpdateChange(currentEndpoint, desiredEndpoint)

		e := &Endpoint{
			Addresses:        ipsToAddresses(ips),
			IPs:              ips,
			Ports:            currentEndpoint.Ports,
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
		}

		updateChange = r.newK8sEndpoint(e)
	}

	return updateChange, nil
}

func ipsForUpdateChange(currentEndpoint *Endpoint, desiredEndpoint *Endpoint) []string {
	var ips []string

	for _, ip := range currentEndpoint.IPs {
		ips = append(ips, ip)
	}

	for _, ip := range desiredEndpoint.IPs {
		if !containsIP(ips, ip) {
			ips = append(ips, ip)
		}
	}

	if len(currentEndpoint.IPs) > 0 {
		return ips
	}

	return nil
}
