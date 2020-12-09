package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	namespaceToCreate, err := toNamespace(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToCreate != nil {
		r.logger.Debugf(ctx, "creating the namespace in the Kubernetes API")

		_, err = r.k8sClient.CoreV1().Namespaces().Create(ctx, namespaceToCreate, v1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "created the namespace in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the namespace does not need to be created in the Kubernetes API")
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

	r.logger.Debugf(ctx, "finding out if the namespace has to be created")

	var namespaceToCreate *corev1.Namespace
	if currentNamespace == nil {
		namespaceToCreate = desiredNamespace
	}

	r.logger.Debugf(ctx, "found out if the namespace has to be created")

	return namespaceToCreate, nil
}
