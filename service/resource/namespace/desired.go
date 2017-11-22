package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogWithCtx(ctx, "debug", "computing the desired namespace")

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

	r.logger.LogWithCtx(ctx, "debug", "computed the desired namespace")

	return namespace, nil
}
