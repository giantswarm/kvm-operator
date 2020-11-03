package clusterrolebinding

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	apiv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			_, err := r.k8sClient.RbacV1().ClusterRoleBindings().Update(ctx, clusterRoleBinding, apismetav1.UpdateOptions{})
			if isExternalFieldImmutableError(err) {
				// We can't change a RoleRef, so delete the old CRB and replace it
				r.logger.LogCtx(ctx, "level", "debug", "message", "unable to update immutable field, re-creating the cluster role binding instead")

				r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the old cluster role binding")
				// Delete the old CRB
				err = r.k8sClient.RbacV1().ClusterRoleBindings().Delete(ctx, clusterRoleBinding.Name, apismetav1.DeleteOptions{})
				if apierrors.IsNotFound(err) {
				} else if err != nil {
					return microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", "creating the new cluster role binding")
				// Create the new CRB
				_, err = r.k8sClient.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, apismetav1.CreateOptions{})
				if apierrors.IsAlreadyExists(err) {
				} else if err != nil {
					return microerror.Mask(err)
				}

			} else if err != nil {
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
