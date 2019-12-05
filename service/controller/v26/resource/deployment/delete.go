package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v26/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToDelete, err := toDeployments(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToDelete) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the deployments in the Kubernetes API")

		for _, deployment := range deploymentsToDelete {
			n := key.ClusterNamespace(customResource)
			err := r.k8sClient.AppsV1().Deployments(n).Delete(deployment.Name, newDeleteOptions())
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the deployments in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the deployments do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

// newDeleteChangeForDeletePatch is used on delete events to get rid of all
// deployments.
func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which deployments have to be deleted")

	var deploymentsToDelete []*v1.Deployment

	for _, currentDeployment := range currentDeployments {
		if containsDeployment(desiredDeployments, currentDeployment) {
			deploymentsToDelete = append(deploymentsToDelete, currentDeployment)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d deployments that have to be deleted", len(deploymentsToDelete)))

	return deploymentsToDelete, nil
}

// newDeleteChangeForUpdatePatch is used on update events to scale down
// deployments.
func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which deployments have to be deleted")

	var deploymentsToDelete []*v1.Deployment

	for _, currentDeployment := range currentDeployments {
		if !containsDeployment(desiredDeployments, currentDeployment) {
			deploymentsToDelete = append(deploymentsToDelete, currentDeployment)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d deployments that have to be deleted", len(deploymentsToDelete)))

	return deploymentsToDelete, nil
}

func newDeleteOptions() *metav1.DeleteOptions {
	propagation := metav1.DeletePropagationForeground

	options := &metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	}

	return options
}
