package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v6/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	serviceAccountToDelete, err := toServiceAccount(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if serviceAccountToDelete != nil {
		r.logger.LogCtx(ctx, "debug", "deleting the service account in the Kubernetes API")

		// Delete service account in the Kubernetes API.
		namespace := key.ClusterNamespace(customObject)
		err := r.k8sClient.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountToDelete.Name, &apismetav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "deleted the service account in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the service account does not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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

	r.logger.LogCtx(ctx, "debug", "finding out which service account has to be deleted")

	var serviceAccountToDelete *apiv1.ServiceAccount
	if currentServiceAccount != nil {
		serviceAccountToDelete = desiredServiceAccount
	}

	r.logger.LogCtx(ctx, "debug", "found out service account that has to be deleted")

	return serviceAccountToDelete, nil
}
