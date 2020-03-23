package endpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	endpointToDelete, err := toK8sEndpoint(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// The endpoint resource is reconciled by watching pods. Pods get deleted at
	// times. We do not want to delete the whole endpoint only because one pod is
	// gone. We only delete the whole endpoint when it does not contain any IP
	// anymore. Removing IPs is done on update events.
	if isEmptyEndpoint(endpointToDelete) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting endpoint '%s'", endpointToDelete.GetName()))

		err = r.k8sClient.CoreV1().Endpoints(endpointToDelete.Namespace).Delete(endpointToDelete.GetName(), &metav1.DeleteOptions{})
		if errors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted endpoint '%s'", endpointToDelete.GetName()))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting endpoint '%s'", endpointToDelete.GetName()))
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(deleteChange)
	patch.SetUpdateChange(deleteChange)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var deleteChange *corev1.Endpoints
	{
		ips := ipsForDeleteChange(currentEndpoint.IPs, desiredEndpoint.IPs)

		e := &Endpoint{
			Addresses:        ipsToAddresses(ips),
			IPs:              ips,
			Ports:            currentEndpoint.Ports,
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
		}

		deleteChange = r.newK8sEndpoint(e)
	}

	return deleteChange, nil
}

func ipsForDeleteChange(currentIPs []string, desiredIPs []string) []string {
	var ips []string

	for _, ip := range currentIPs {
		if !containsIP(ips, ip) {
			ips = append(ips, ip)
		}
	}

	for _, ip := range desiredIPs {
		if containsIP(ips, ip) {
			ips = removeIP(ips, ip)
		}
	}

	return ips
}
