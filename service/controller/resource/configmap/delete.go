package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMMachine(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.k8sClient.CtrlClient().Delete(ctx, &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.ConfigMapName(cr),
			Namespace: key.ClusterNamespace(&cr),
		},
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
