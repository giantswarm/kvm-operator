package serviceaccount

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/kvm-operator/service/controller/v20/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	serviceAccountToCreate, err := toServiceAccount(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create the service account in the Kubernetes API.
	if serviceAccountToCreate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the service account in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		_, err := r.k8sClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccountToCreate)
		if apierrors.IsAlreadyExists(err) {
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the service account in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the service account does not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentServiceAccount, err := toServiceAccount(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServiceAccount, err := toServiceAccount(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which service account has to be created")

	var serviceAccountToCreate *apiv1.ServiceAccount
	if currentServiceAccount == nil {
		serviceAccountToCreate = desiredServiceAccount
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found out that service account that has to be created"))

	return serviceAccountToCreate, nil
}
