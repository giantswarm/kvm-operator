package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	namespaceToDelete, err := toNamespace(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToDelete != nil {
		r.logger.Debugf(ctx, "deleting the namespace in the Kubernetes API")

		err = r.ctrlClient.Delete(ctx, namespaceToDelete)
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted the namespace in the Kubernetes API")
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.Debugf(ctx, "canceling reconciliation")
	} else {
		r.logger.Debugf(ctx, "the namespace does not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := toNamespace(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out if the namespace has to be deleted")

	var namespaceToDelete *corev1.Namespace
	if currentNamespace != nil {
		namespaceToDelete = desiredNamespace
	}

	r.logger.Debugf(ctx, "found out if the namespace has to be deleted")

	return namespaceToDelete, nil
}
