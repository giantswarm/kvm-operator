package endpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/kvm-operator/service/controller/v15/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
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
	if isEmptyEndpoint(*endpointToDelete) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting endpoint '%s'", endpointToDelete.GetName()))

		err = r.k8sClient.CoreV1().Endpoints(endpointToDelete.Namespace).Delete(endpointToDelete.Name, &metav1.DeleteOptions{})
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

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(deleteChange)
	patch.SetUpdateChange(deleteChange)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding current version of the reconciled pod in the Kubernetes API")

		currentPod, err := r.k8sClient.CoreV1().Pods(reconciledPod.GetNamespace()).Get(reconciledPod.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// In case we reconcile a pod we cannot find anymore this means the
			// informer's watch event is outdated and the pod got already deleted in
			// the Kubernetes API. This is a normal transition behaviour, so we just
			// ignore it and continue with endpoint deletion.
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find current version of the reconciled pod in the Kubernetes API")
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the current version of the reconciled pod in the Kubernetes API")

			if !key.ArePodContainersTerminated(currentPod) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "pod containers are still running")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)

				return nil, nil
			}
		}
	}

	var deleteChange *corev1.Endpoints
	{
		endpoint := &Endpoint{
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
			IPs:              cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs),
		}
		deleteChange, err = r.newK8sEndpoint(endpoint)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return deleteChange, nil
}
