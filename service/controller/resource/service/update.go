package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	servicesToUpdate, err := toServices(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(servicesToUpdate) > 0 {
		r.logger.Debugf(ctx, "updating services")

		for _, serviceToUpdate := range servicesToUpdate {
			_, err := r.k8sClient.CoreV1().Services(serviceToUpdate.Namespace).Update(ctx, serviceToUpdate, metav1.UpdateOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "updated services")
	} else {
		r.logger.Debugf(ctx, "no need to update services")
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
	currentServices, err := toServices(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServices, err := toServices(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which services have to be updated")

	servicesToUpdate := make([]*corev1.Service, 0)

	for _, currentService := range currentServices {
		desiredService, err := getServiceByName(desiredServices, currentService.Name)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isServiceModified(desiredService, currentService) {
			latest, err := r.k8sClient.CoreV1().Services(desiredService.GetNamespace()).Get(ctx, desiredService.GetName(), metav1.GetOptions{})
			if err != nil {
				return nil, microerror.Mask(err)
			}

			serviceToUpdate := desiredService.DeepCopy()
			serviceToUpdate.ObjectMeta.ResourceVersion = latest.GetResourceVersion()
			serviceToUpdate.Spec.ClusterIP = currentService.Spec.ClusterIP

			servicesToUpdate = append(servicesToUpdate, serviceToUpdate)

			r.logger.Debugf(ctx, "found service '%s' that has to be updated", desiredService.GetName())
		}
	}

	r.logger.Debugf(ctx, "found %d services which have to be updated", len(servicesToUpdate))

	return servicesToUpdate, nil
}
