package deploymentv3

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"k8s.io/api/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv3"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := keyv3.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToUpdate, err := toDeployments(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "debug", "updating the deployments in the Kubernetes API")

		namespace := keyv3.ClusterNamespace(customObject)
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

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
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

	patch := framework.NewPatch()
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

		// Updates can be quite disruptive. We have to be very careful with updating
		// resources that potentially imply disrupting customer workloads. We have
		// to check the state of all deployments before we can safely go ahead with
		// the update procedure.
		for _, d := range currentDeployments {
			allReplicasUp := allNumbersEqual(d.Status.AvailableReplicas, d.Status.ReadyReplicas, d.Status.Replicas, d.Status.UpdatedReplicas)
			if !allReplicasUp {
				r.logger.LogCtx(ctx, "warning", fmt.Sprintf("cannot update any deployment: deployment '%s' must have all replicas up", d.GetName(), VersionBundleVersionAnnotation))
				return nil, nil
			}
		}

		// We select one deployment to be updated per reconciliation loop. Therefore
		// we have to check its state on the version bundle level to see if a
		// deployment is already up to date. We also check if there are any other
		// changes on the pod specs. In case there are not, we check the next one.
		// The first one not being up to date will be chosen to be updated next and
		// the loop will be broken immediatelly.
		for _, currentDeployment := range currentDeployments {
			desiredDeployment, err := getDeploymentByName(desiredDeployments, currentDeployment.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			isModified, err := isDeploymentModified(desiredDeployment, currentDeployment)
			if IsMissingAnnotation(err) {
				r.logger.LogCtx(ctx, "warning", fmt.Sprintf("cannot update current deployment '%s': annotation '%s' must not be missing", currentDeployment.GetName(), VersionBundleVersionAnnotation))
			} else if IsEmptyAnnotation(err) {
				r.logger.LogCtx(ctx, "warning", fmt.Sprintf("cannot update current deployment '%s': annotation '%s' must not be empty", currentDeployment.GetName(), VersionBundleVersionAnnotation))
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if !isModified {
				r.logger.LogCtx(ctx, "warning", fmt.Sprintf("not updating deployment '%s': no changes found", currentDeployment.GetName()))
				continue
			}

			deploymentsToUpdate = append(deploymentsToUpdate, desiredDeployment)
			break
		}

		r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d deployment(s) that have to be updated", len(deploymentsToUpdate)))
	} else {
		r.logger.LogCtx(ctx, "debug", "not computing update state because deployments are not allowed to be updated")
	}

	return deploymentsToUpdate, nil
}
