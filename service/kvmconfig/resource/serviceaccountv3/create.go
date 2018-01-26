package serviceaccountv3

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/keyv3"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := keyv3.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	serviceAccountToCreate, err := toServiceAccount(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create the service account in the Kubernetes API.
	if serviceAccountToCreate != nil {
		r.logger.LogCtx(ctx, "debug", "creating the service account in the Kubernetes API")

		namespace := keyv3.ClusterNamespace(customObject)
		_, err := r.k8sClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccountToCreate)
		if apierrors.IsAlreadyExists(err) {
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "created the service account in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the service account does not need to be created in the Kubernetes API")
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

	r.logger.LogCtx(ctx, "debug", "finding out which service account has to be created")

	var serviceAccountToCreate *apiv1.ServiceAccount
	if currentServiceAccount == nil {
		serviceAccountToCreate = desiredServiceAccount
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found out that service account that has to be created"))

	return serviceAccountToCreate, nil
}
