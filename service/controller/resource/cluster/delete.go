package cluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		cluster := v1alpha2.KVMCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.ClusterID(&cr),
				Namespace: cr.Namespace,
			},
		}
		err = r.ctrlClient.Delete(ctx, &cluster)
		if err != nil && !apierrors.IsNotFound(err) {
			return microerror.Mask(err)
		}
	}

	{
		cluster := capiv1alpha3.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.ClusterID(&cr),
				Namespace: cr.Namespace,
			},
		}
		err = r.ctrlClient.Delete(ctx, &cluster)
		if err != nil && !apierrors.IsNotFound(err) {
			return microerror.Mask(err)
		}
	}

	return nil
}
