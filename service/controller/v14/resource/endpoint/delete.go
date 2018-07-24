package endpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if k8sEndpoint == nil {
		return nil // Nothing to do.
	}

	// TODO this looks very wrong. When the endpoint is not empty on a delete
	// event, we want to delete it usually. The code below says we don't.
	if !isEmptyEndpoint(*k8sEndpoint) {
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting endpoint '%s'", k8sEndpoint.GetName()))

	err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Delete(k8sEndpoint.Name, &metav1.DeleteOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted endpoint '%s'", k8sEndpoint.GetName()))

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	deleteState, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	updateState, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(deleteState)
	patch.SetUpdateChange(updateState)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	ips := cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs)

	if currentEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if desiredEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if len(ips) > 0 {
		return nil, nil
	}

	endpoint := &Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
		IPs:              ips,
	}
	deleteState, err := r.newK8sEndpoint(endpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return deleteState, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	ips := cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs)

	if currentEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if desiredEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if len(ips) == 0 {
		return nil, nil
	}

	endpoint := &Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
		IPs:              ips,
	}
	updateState, err := r.newK8sEndpoint(endpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return updateState, nil
}
