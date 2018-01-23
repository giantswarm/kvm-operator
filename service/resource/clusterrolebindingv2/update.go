package clusterrolebindingv2

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/rbac/v1beta1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	clusterRoleBindingsToUpdate, err := toClusterRoleBindings(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(clusterRoleBindingsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "debug", "updating the cluster role bindings in the Kubernetes API")

		// Create the cluster role bindings in the Kubernetes API.
		for _, clusterRoleBinding := range clusterRoleBindingsToUpdate {
			_, err := r.k8sClient.RbacV1beta1().ClusterRoleBindings().Update(clusterRoleBinding)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "debug", "updated the cluster role bindings in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the cluster role bindings do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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
		r.logger.LogCtx(ctx, "debug", "finding out which cluster role bindings have to be updated")

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

		r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d cluster role bindings that have to be updated", len(clusterRoleBindingsToUpdate)))
	}

	return clusterRoleBindingsToUpdate, nil
}
