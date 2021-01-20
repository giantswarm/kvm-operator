package machine

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

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, node := range cr.Spec.Cluster.Nodes {
		err = r.ensureMachineDeleted(ctx, cr, node)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) ensureMachineDeleted(ctx context.Context, cluster v1alpha2.KVMCluster, node v1alpha2.KVMClusterSpecClusterNode) error {
	machineKey := client.ObjectKey{
		Namespace: key.ClusterNamespace(&cluster),
		Name:      node.ID,
	}

	{
		machine := v1alpha2.KVMMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machineKey.Name,
				Namespace: machineKey.Namespace,
			},
		}
		err := r.ctrlClient.Delete(ctx, &machine)
		if err != nil && !apierrors.IsNotFound(err) {
			return microerror.Mask(err)
		}
	}

	{
		machine := capiv1alpha3.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machineKey.Name,
				Namespace: machineKey.Namespace,
			},
		}
		err := r.ctrlClient.Delete(ctx, &machine)
		if err != nil && !apierrors.IsNotFound(err) {
			return microerror.Mask(err)
		}
	}

	return nil
}
