package configmap

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMMachine(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := cr.Namespace
	var kvmCluster v1alpha2.KVMCluster
	{
		err := r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
			Namespace: clusterID,
			Name:      clusterID,
		}, &kvmCluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	prefix := key.WorkerID
	if cr.Labels["cluster.x-k8s.io/control-plane"] == "true" {
		prefix = key.MasterID
	}

	err = r.k8sClient.CtrlClient().Delete(ctx, &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.ConfigMapName(cr, prefix),
			Namespace: key.ClusterNamespace(&cr),
		},
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
