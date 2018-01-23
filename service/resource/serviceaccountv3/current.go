package serviceaccountv3

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/kvm-operator/service/keyv3"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv3.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for a service account in the Kubernetes API")

	namespace := keyv3.ClusterNamespace(customObject)
	var currentServiceAccount *apiv1.ServiceAccount
	currentServiceAccount, err = r.k8sClient.CoreV1().ServiceAccounts(namespace).Get(keyv3.ServiceAccountName(customObject), apismetav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "debug", "did not find the service account in the Kubernetes API")
		//When service account is not found api still returning non nil value so it can break create/update/delete
		// and is why force the return value to nil
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "debug", "found a service account in the Kubernetes API")
	}

	return currentServiceAccount, nil
}
