package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v20/key"
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the service account in the Kubernetes API")

		// Delete service account in the Kubernetes API.
		namespace := key.ClusterNamespace(customObject)
		err := r.k8sClient.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountToDelete.Name, &metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the service account in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the service account does not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
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

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which service account has to be deleted")

	var serviceAccountToDelete *corev1.ServiceAccount
	if currentServiceAccount != nil {
		serviceAccountToDelete = desiredServiceAccount
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found out service account that has to be deleted")

	return serviceAccountToDelete, nil
}
