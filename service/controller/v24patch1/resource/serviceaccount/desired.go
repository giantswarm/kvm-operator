package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
>>>>>>> d6f149c2... wire v24patch1
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new service account")

	serviceAccount := &apiv1.ServiceAccount{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      key.ServiceAccountName(customObject),
			Namespace: key.ClusterID(customObject),
			Labels: map[string]string{
				"cluster-id":  key.ClusterID(customObject),
				"customer-id": key.ClusterCustomer(customObject),
			},
		},
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed the new service account")

	return serviceAccount, nil
}
