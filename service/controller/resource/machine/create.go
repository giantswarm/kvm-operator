package machine

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
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
		label.OperatorVersion:           project.Version(),
		label.Cluster:                   key.ClusterID(&cluster),
		label.ManagedBy:                 project.Name(),
		label.Organization:              key.ClusterCustomer(&cluster),
		label.ReleaseVersion:            key.ReleaseVersion(&cluster),
	}
	if node.Role == key.MasterID {
		labels[label.ControlPlane] = "true"
	}

	{
		var existing capiv1alpha3.Machine
		err := r.ctrlClient.Get(ctx, machineKey, &existing)
		if apierrors.IsNotFound(err) {
			toCreate := capiv1alpha3.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machineKey.Name,
					Namespace: machineKey.Namespace,
					Labels:    labels,
				},
				Spec: capiv1alpha3.MachineSpec{
					ClusterName: key.ClusterID(&cluster),
					Bootstrap: capiv1alpha3.Bootstrap{
						DataSecretName: to.StringP(fmt.Sprintf("%s-%s-%s", node.Role, key.ClusterID(&cluster), node.ID)),
					},
					InfrastructureRef: corev1.ObjectReference{
						Kind:      node.InfrastructureRef.Kind,
						Namespace: key.ClusterNamespace(&cluster),
						Name:      node.InfrastructureRef.Name,
					},
					ProviderID: to.StringP(fmt.Sprintf("kvm://%s", node.ID)),
				},
				Status: capiv1alpha3.MachineStatus{},
			}
			err = r.ctrlClient.Create(ctx, &toCreate)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		var existing v1alpha2.KVMMachine
		err := r.ctrlClient.Get(ctx, machineKey, &existing)
		if apierrors.IsNotFound(err) {
			toCreate := v1alpha2.KVMMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machineKey.Name,
					Namespace: machineKey.Namespace,
					Labels:    labels,
				},
				Spec: v1alpha2.KVMMachineSpec{
					ProviderID: fmt.Sprintf("kvm://%s", node.ID),
					Size:       node.Size,
				},
			}
			err = r.ctrlClient.Create(ctx, &toCreate)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}

		needsUpdate := !reflect.DeepEqual(node.Size, existing.Spec.Size)
		if needsUpdate {
			existing.Spec.Size = node.Size
			err = r.ctrlClient.Update(ctx, &existing)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}
