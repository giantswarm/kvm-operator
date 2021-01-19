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

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, master := range cr.Spec.Cluster.Masters {
		err = r.ensureMachineCreated(ctx, cr, key.DeploymentName(key.MasterID, master.ID))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, worker := range cr.Spec.Cluster.Workers {
		err = r.ensureMachineCreated(ctx, cr, key.DeploymentName(key.WorkerID, worker.ID))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) ensureMachineCreated(ctx context.Context, cr v1alpha2.KVMCluster, name string) error {
	{
		var cluster capiv1alpha3.Machine
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: key.ClusterNamespace(cr),
			Name:      name,
		}, &cluster)
		if apierrors.IsNotFound(err) {
			cluster = capiv1alpha3.Machine{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       capiv1alpha3.MachineSpec{},
				Status:     capiv1alpha3.MachineStatus{},
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
		var kvmMachine v1alpha2.KVMMachine
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: key.ClusterNamespace(cr),
			Name:      name,
		}, &kvmMachine)
		if apierrors.IsNotFound(err) {
			kvmMachine = v1alpha2.KVMMachine{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1alpha2.KVMMachineSpec{},
				Status:     v1alpha2.KVMMachineStatus{},
			}
			err = r.ctrlClient.Create(ctx, &kvmMachine)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
