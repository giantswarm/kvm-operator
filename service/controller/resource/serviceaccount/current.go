package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if key.IsDeleted(&customObject) {
		r.logger.Debugf(ctx, "redirecting responsibility of deletion of service accounts to namespace termination")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.Debugf(ctx, "canceling resource")

		return nil, nil
	}

	r.logger.Debugf(ctx, "looking for a service account in the Kubernetes API")

	namespace := key.ClusterNamespace(customObject)
	var currentServiceAccount corev1.ServiceAccount
	err = r.ctrlClient.Get(ctx, client.ObjectKey{
		Name:      key.ServiceAccountName(customObject),
		Namespace: namespace,
	}, &currentServiceAccount)
	if apierrors.IsNotFound(err) {
		r.logger.Debugf(ctx, "did not find the service account in the Kubernetes API")
		//When service account is not found api still returning non nil value so it can break create/update/delete
		// and is why force the return value to nil
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		r.logger.Debugf(ctx, "found a service account in the Kubernetes API")
	}

	return &currentServiceAccount, nil
}
