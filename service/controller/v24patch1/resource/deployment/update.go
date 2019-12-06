package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/resource/crud"
	v1 "k8s.io/api/apps/v1"

<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToUpdate, err := toDeployments(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customResource)
		for _, deployment := range deploymentsToUpdate {
			_, err := r.k8sClient.AppsV1().Deployments(namespace).Update(deployment)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated the deployments in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the deployments do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
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

	patch := crud.NewPatch()
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

	if updateallowedcontext.IsUpdateAllowed(ctx) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which deployments have to be updated")

		// Updates can be quite disruptive. We have to be very careful with updating
		// resources that potentially imply disrupting customer workloads. We have
		// to check the state of all deployments before we can safely go ahead with
		// the update procedure.
		for _, d := range currentDeployments {
			allReplicasUp := allNumbersEqual(d.Status.AvailableReplicas, d.Status.ReadyReplicas, d.Status.Replicas, d.Status.UpdatedReplicas)
			if !allReplicasUp {
				r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("cannot update any deployment: deployment '%s' must have all replicas up", d.GetName()))
				return nil, nil
			}
		}

		// We select one deployment to be updated per reconciliation loop. Therefore
		// we have to check its state on the version bundle level to see if a
		// deployment is already up to date. We also check if there are any other
		// changes on the pod specs. In case there are none, we check the next one.
		// The first one not being up to date will be chosen to be updated next and
		// the loop will be broken immediatelly.
		for _, currentDeployment := range currentDeployments {
			desiredDeployment, err := getDeploymentByName(desiredDeployments, currentDeployment.Name)
			if IsNotFound(err) {
				// NOTE that this case indicates we should remove the current deployment
				// eventually.
				r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("not updating deployment '%s': no desired deployment found", currentDeployment.GetName()))
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if !isDeploymentModified(desiredDeployment, currentDeployment) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not updating deployment '%s': no changes found", currentDeployment.GetName()))
				continue
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found deployment '%s' that has to be updated", desiredDeployment.GetName()))

			return []*v1.Deployment{desiredDeployment}, nil
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing update state because deployments are not allowed to be updated")
	}

	return nil, nil
}
