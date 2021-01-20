package serviceaccount

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computing the new service account")

	serviceAccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.ServiceAccountName(cr),
			Namespace: key.ClusterID(&cr),
			Labels: map[string]string{
				"cluster-id":  key.ClusterID(&cr),
				"customer-id": key.ClusterCustomer(&cr),
			},
		},
	}

	r.logger.Debugf(ctx, "computed the new service account")

	return serviceAccount, nil
}
