package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	namespaceToCreate, err := toNamespace(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToCreate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the namespace in the Kubernetes API")

		_, err = r.k8sClient.CoreV1().Namespaces().Create(namespaceToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the namespace in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the namespace does not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := toNamespace(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the namespace has to be created")

	var namespaceToCreate *apiv1.Namespace
	if currentNamespace == nil {
		namespaceToCreate = desiredNamespace
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found out if the namespace has to be created")

	return namespaceToCreate, nil
}
