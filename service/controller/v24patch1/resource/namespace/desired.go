package namespace

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

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the desired namespace")

	// Compute the desired state of the namespace to have a reference of data how
	// it should be.
	namespace := &apiv1.Namespace{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.ClusterNamespace(customObject),
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
			},
		},
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed the desired namespace")

	return namespace, nil
}
