package serviceaccountv3

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/keyv3"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv3.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "computing the new service account")

	serviceAccount := &apiv1.ServiceAccount{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      keyv3.ServiceAccountName(customObject),
			Namespace: keyv3.ClusterID(customObject),
			Labels: map[string]string{
				"cluster-id":  keyv3.ClusterID(customObject),
				"customer-id": keyv3.ClusterCustomer(customObject),
			},
		},
	}

	r.logger.LogCtx(ctx, "debug", "computed the new service account")

	return serviceAccount, nil
}
