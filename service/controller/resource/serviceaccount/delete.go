package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	serviceAccountToDelete, err := toServiceAccount(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if serviceAccountToDelete != nil {
		r.logger.Debugf(ctx, "deleting the service account in the Kubernetes API")

		err := r.ctrlClient.Delete(ctx, serviceAccountToDelete)
		if apierrors.IsNotFound(err) {
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted the service account in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the service account does not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(deleteChange)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentServiceAccount, err := toServiceAccount(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServiceAccount, err := toServiceAccount(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which service account has to be deleted")

	var serviceAccountToDelete *corev1.ServiceAccount
	if currentServiceAccount != nil {
		serviceAccountToDelete = desiredServiceAccount
	}

	r.logger.Debugf(ctx, "found out service account that has to be deleted")

	return serviceAccountToDelete, nil
}
