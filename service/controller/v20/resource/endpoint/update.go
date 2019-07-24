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
		ips := ipsForUpdateChange(currentEndpoint.IPs, desiredEndpoint.IPs)

		e := &Endpoint{
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
		}

		if !ipsAreEqual(currentEndpoint.IPs, desiredEndpoint.IPs) {
			e.Addresses = ipsToAddresses(ips)
			e.IPs = ips
			e.Ports = currentEndpoint.Ports
		}

		updateChange = r.newK8sEndpoint(e)
	}

	return updateChange, nil
}

func ipsAreEqual(currentIPs []string, desiredIPs []string) bool {
	// In case one slice is nil and the other is not, it is not equal anymore.
	if (currentIPs == nil) != (desiredIPs == nil) {
		return false
	}

	// In case one slice has more or less items in it, it is not equal anymore.
	if len(currentIPs) != len(desiredIPs) {
		return false
	}

	// In case one slice is missing some item, it is not equal anymore.
	for i := range currentIPs {
		if currentIPs[i] != desiredIPs[i] {
			return false
		}
	}

	return true
}

func ipsForUpdateChange(currentIPs []string, desiredIPs []string) []string {
	var ips []string

	for _, ip := range currentIPs {
		ips = append(ips, ip)
	}

	for _, ip := range desiredIPs {
		if !containsIP(ips, ip) {
			ips = append(ips, ip)
		}
	}

	if len(currentIPs) > 0 {
		return ips
	}

	return nil
}
