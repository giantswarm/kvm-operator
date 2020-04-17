package clusterrolebinding

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
	apiv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	clusterRoleBindingsToUpdate, err := toClusterRoleBindings(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(clusterRoleBindingsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating the cluster role bindings in the Kubernetes API")

		// Create the cluster role bindings in the Kubernetes API.
		for _, clusterRoleBinding := range clusterRoleBindingsToUpdate {
			_, err := r.k8sClient.RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
			r.logger.Log("level", "debug", "message", fmt.Sprintf(" got error: %v#", err))
			r.logger.Log("level", "debug", "message", fmt.Sprintf(" error(): %v#", err.Error()))
			if isExternalFieldImmutableError(err) {
				// Create new CRB and delete the old one
				r.logger.Log("level", "debug", "message", "Should re-create CRB")

				// Simply create this as a new CRB
				newCRB := apiv1.ClusterRoleBinding{}
				clusterRoleBinding.DeepCopyInto(&newCRB)
				newCRB.SetName(clusterRoleBinding.Name + "-upgrading")

				r.logger.Log("level", "debug", "message", "Creating new CRB instead of updating")
				r.logger.Log("level", "debug", "message", fmt.Sprintf("Creating CRB: %v#", newCRB))
				_, err := r.k8sClient.RbacV1().ClusterRoleBindings().Create(&newCRB)
				if apierrors.IsAlreadyExists(err) {
				} else if err != nil {
					r.logger.Log("level", "debug", "message", "This error is from creation")
					return microerror.Mask(err)
				}

				// Get existing CRB

				// Change its name

				// Create the new one using the original name

				// Delete the renamed one. Wait for all old Pods to be deleted?

				// Alternatively, create the new one with a new name
			} else if err != nil {
				r.logger.Log("level", "debug", "message", "This error is not an immutable error")
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated the cluster role bindings in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the cluster role bindings do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentClusterRoleBindings, err := toClusterRoleBindings(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredClusterRoleBindings, err := toClusterRoleBindings(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusterRoleBindingsToUpdate []*apiv1.ClusterRoleBinding
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which cluster role bindings have to be updated")

		for _, clusterRoleBinding := range currentClusterRoleBindings {
			desiredClusterRoleBinding, err := getClusterRoleBindingByName(desiredClusterRoleBindings, clusterRoleBinding.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if isClusterRoleBindingModified(desiredClusterRoleBinding, clusterRoleBinding) {
				clusterRoleBindingsToUpdate = append(clusterRoleBindingsToUpdate, desiredClusterRoleBinding)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d cluster role bindings that have to be updated", len(clusterRoleBindingsToUpdate)))
	}

	return clusterRoleBindingsToUpdate, nil
}
