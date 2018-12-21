package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v18/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	servicesToDelete, err := toServices(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(servicesToDelete) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the services in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, service := range servicesToDelete {
			err := r.k8sClient.CoreV1().Services(namespace).Delete(service.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the services in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the services do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentServices, err := toServices(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServices, err := toServices(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which services have to be deleted")

	var servicesToDelete []*apiv1.Service

	for _, currentService := range currentServices {
		if containsService(desiredServices, currentService) {
			servicesToDelete = append(servicesToDelete, currentService)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d services that have to be deleted", len(servicesToDelete)))

	return servicesToDelete, nil
}
