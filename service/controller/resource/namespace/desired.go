package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToKVMConfig(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computing the desired namespace")

	// Compute the desired state of the namespace to have a reference of data how
	// it should be.
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Spec.Cluster.ID,
			Labels: map[string]string{
				"cluster":  cr.Spec.Cluster.ID,
				"customer": cr.Spec.Cluster.Customer.ID,
			},
		},
	}

	r.logger.Debugf(ctx, "computed the desired namespace")

	return namespace, nil
}
