package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"k8s.io/api/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/controller/v2/key"
	"github.com/giantswarm/kvm-operator/service/controller/v2/messagecontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToUpdate, err := toDeployments(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "debug", "updating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToUpdate {
			_, err := r.k8sClient.Extensions().Deployments(namespace).Update(deployment)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "debug", "updated the deployments in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the deployments do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var deploymentsToUpdate []*v1beta1.Deployment
	if updateallowedcontext.IsUpdateAllowed(ctx) {
		r.logger.LogCtx(ctx, "debug", "finding out which deployments have to be updated")

		// Check if config maps of deployments changed. In case they did, add the
		// deployments to the list of deployments intended to be updated.
		m, ok := messagecontext.FromContext(ctx)
		if ok {
			for _, name := range m.ConfigMapNames {
				desiredDeployment, err := getDeploymentByConfigMapName(desiredDeployments, name)
				if err != nil {
					return nil, microerror.Mask(err)
				}
				deploymentsToUpdate = append(deploymentsToUpdate, desiredDeployment)
			}
		}

		// Check if deployments changed. In case they did, add the deployments to
		// the list of deployments intended to be updated, but only in case they are
		// not already being tracked.
		for _, currentDeployment := range currentDeployments {
			desiredDeployment, err := getDeploymentByName(desiredDeployments, currentDeployment.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if !isDeploymentModified(desiredDeployment, currentDeployment) {
				continue
			}

			if containsDeployment(deploymentsToUpdate, desiredDeployment) {
				continue
			}

			deploymentsToUpdate = append(deploymentsToUpdate, desiredDeployment)
		}

		r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d deployments that have to be updated", len(deploymentsToUpdate)))
	} else {
		r.logger.LogCtx(ctx, "debug", "not computing update state because deployments are not allowed to be updated")
	}

	return deploymentsToUpdate, nil
}
