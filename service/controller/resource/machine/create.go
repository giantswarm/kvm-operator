package machine

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	corev1 "k8s.io/api/core/v1"
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

	for _, node := range cr.Spec.Cluster.Nodes {
		err = r.ensureMachineCreated(ctx, cr, node)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) ensureMachineCreated(ctx context.Context, cluster v1alpha2.KVMCluster, node v1alpha2.KVMClusterSpecClusterNode) error {
	machineKey := client.ObjectKey{
		Namespace: key.ClusterNamespace(&cluster),
		Name:      node.ID,
	}

	labels := map[string]string{
		"cluster.x-k8s.io/cluster-name": key.ClusterID(&cluster),
	}
	if node.Role == "master" {
		labels["cluster.x-k8s.io/control-plane"] = "true"
	}

	{
		var cluster capiv1alpha3.Machine
		err := r.ctrlClient.Get(ctx, machineKey, &cluster)
		if apierrors.IsNotFound(err) {
			cluster = capiv1alpha3.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machineKey.Name,
					Namespace: machineKey.Namespace,
					Labels:    labels,
				},
				Spec: capiv1alpha3.MachineSpec{
					ClusterName: key.ClusterID(&cluster),
					Bootstrap: capiv1alpha3.Bootstrap{
						DataSecretName: nil,
					},
					InfrastructureRef: corev1.ObjectReference{
						Kind:      node.InfrastructureRef.Kind,
						Namespace: key.ClusterNamespace(&cluster),
						Name:      node.InfrastructureRef.Name,
					},
					ProviderID: to.StringP(node.ID),
				},
				Status: capiv1alpha3.MachineStatus{},
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
		err := r.ctrlClient.Get(ctx, machineKey, &kvmMachine)
		if apierrors.IsNotFound(err) {
			kvmMachine = v1alpha2.KVMMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machineKey.Name,
					Namespace: machineKey.Namespace,
					Labels:    labels,
				},
				Spec: v1alpha2.KVMMachineSpec{
					ProviderID: node.ID,
					Size:       node.Size,
				},
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
