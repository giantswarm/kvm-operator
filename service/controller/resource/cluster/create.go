package cluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		var cluster capiv1alpha3.Cluster
		err = r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: cr.Namespace,
			Name:      key.ClusterID(cr),
		}, &cluster)
		if apierrors.IsNotFound(err) {
			cluster = capiv1alpha3.Cluster{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       capiv1alpha3.ClusterSpec{},
				Status:     capiv1alpha3.ClusterStatus{},
			}
			err = r.ctrlClient.Create(ctx, &cluster)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		var kvmCluster v1alpha2.KVMCluster
		err = r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: cr.Namespace,
			Name:      key.ClusterID(cr),
		}, &kvmCluster)
		if apierrors.IsNotFound(err) {
			kvmCluster = v1alpha2.KVMCluster{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1alpha2.KVMClusterSpec{},
				Status:     v1alpha2.KVMClusterStatus{},
			}
			err = r.ctrlClient.Create(ctx, &kvmCluster)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
