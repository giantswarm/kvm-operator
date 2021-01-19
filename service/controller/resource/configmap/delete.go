package configmap

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
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
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: clusterID,
			Name:      clusterID,
		}, &kvmCluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var node v1alpha1.ClusterNode
	role := cr.Spec.ProviderID
	prefix := key.WorkerID
	if role == "master" {
		prefix = key.MasterID
	}

	err = r.ctrlClient.Delete(ctx, &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.ConfigMapName(kvmCluster, node, prefix),
			Namespace: key.ClusterNamespace(kvmCluster),
		},
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
