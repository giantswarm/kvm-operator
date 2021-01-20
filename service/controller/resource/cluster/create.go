package cluster

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
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
	cr, err := key.ToKVMConfig(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	endpoint := capiv1alpha3.APIEndpoint{
		Host: cr.Spec.Cluster.Kubernetes.API.Domain,
		Port: int32(cr.Spec.Cluster.Kubernetes.API.SecurePort),
	}

	clusterKey := client.ObjectKey{
		Namespace: cr.Spec.Cluster.ID,
		Name:      cr.Spec.Cluster.ID,
	}

	{
		var cluster capiv1alpha3.Cluster
		err = r.ctrlClient.Get(ctx, clusterKey, &cluster)

		if apierrors.IsNotFound(err) {
			cluster = capiv1alpha3.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterKey.Name,
					Namespace: clusterKey.Namespace,
					Labels: map[string]string{
						label.OperatorVersion: project.Version(),
						label.Cluster:         cr.Spec.Cluster.ID,
						label.ManagedBy:       project.Name(),
						label.Organization:    cr.Spec.Cluster.Customer.ID,
						label.ReleaseVersion:  key.ReleaseVersion(&cr),
					},
				},
				Spec: capiv1alpha3.ClusterSpec{
					Paused:               false,
					ControlPlaneEndpoint: endpoint,
				},
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
		var nodes []v1alpha2.KVMClusterSpecClusterNode
		for i, master := range cr.Spec.Cluster.Masters {
			nodes = append(nodes, v1alpha2.KVMClusterSpecClusterNode{
				ID:   master.ID,
				Role: key.MasterID,
				InfrastructureRef: corev1.TypedLocalObjectReference{
					APIGroup: to.StringP("infrastructure.giantswarm.io"),
					Kind:     v1alpha2.KindKVMMachine,
					Name:     master.ID,
				},
				Size: v1alpha2.KVMMachineSpecSize(cr.Spec.KVM.Masters[i]),
			})
		}
		for i, worker := range cr.Spec.Cluster.Workers {
			nodes = append(nodes, v1alpha2.KVMClusterSpecClusterNode{
				ID:   worker.ID,
				Role: key.WorkerID,
				InfrastructureRef: corev1.TypedLocalObjectReference{
					APIGroup: to.StringP(v1alpha2.SchemeGroupVersion.Group),
					Kind:     v1alpha2.KindKVMMachine,
					Name:     worker.ID,
				},
				Size: v1alpha2.KVMMachineSpecSize(cr.Spec.KVM.Workers[i]),
			})
		}

		var kvmCluster v1alpha2.KVMCluster
		err = r.ctrlClient.Get(ctx, clusterKey, &kvmCluster)

		if apierrors.IsNotFound(err) {
			var portMappings []v1alpha2.KVMClusterSpecProviderPortMappings
			for _, portMapping := range cr.Spec.KVM.PortMappings {
				portMappings = append(portMappings, v1alpha2.KVMClusterSpecProviderPortMappings(portMapping))
			}

			var description string
			{
				var clusterConfig v1alpha1.KVMClusterConfig
				err = r.ctrlClient.Get(ctx, client.ObjectKey{
					Name:      fmt.Sprintf("%s-%s", cr.Spec.Cluster.ID, "kvm-cluster-config"),
					Namespace: metav1.NamespaceDefault,
				}, &clusterConfig)
				if err != nil {
					return microerror.Mask(err)
				}
				description = clusterConfig.Spec.Guest.Name
			}

			kvmCluster = v1alpha2.KVMCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterKey.Name,
					Namespace: clusterKey.Namespace,
					Labels: map[string]string{
						label.OperatorVersion: project.Version(),
						label.Cluster:         cr.Spec.Cluster.ID,
						label.ManagedBy:       project.Name(),
						label.Organization:    cr.Spec.Cluster.Customer.ID,
						label.ReleaseVersion:  key.ReleaseVersion(&cr),
					},
				},
				Spec: v1alpha2.KVMClusterSpec{
					ControlPlaneEndpoint: endpoint,
					Cluster: v1alpha2.KVMClusterSpecCluster{
						Description: description,
						DNS: v1alpha2.KVMClusterSpecClusterDNS{
							Domain: cr.Spec.Cluster.Kubernetes.Domain,
						},
						Nodes: nodes,
					},
					Provider: v1alpha2.KVMClusterSpecProvider{
						EndpointUpdaterImage: cr.Spec.KVM.EndpointUpdater.Docker.Image,
						MachineImage:         cr.Spec.KVM.K8sKVM.Docker.Image,
						MachineStorageType:   cr.Spec.KVM.K8sKVM.StorageType,
						FlannelVNI:           cr.Spec.KVM.Network.Flannel.VNI,
						PortMappings:         portMappings,
					},
				},
			}
			err = r.ctrlClient.Create(ctx, &kvmCluster)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}

		needsUpdate := !reflect.DeepEqual(nodes, kvmCluster.Spec.Cluster.Nodes)
		if needsUpdate {
			kvmCluster.Spec.Cluster.Nodes = nodes
			err = r.ctrlClient.Update(ctx, &kvmCluster)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		statusNeedsUpdate := !reflect.DeepEqual(cr.Status.KVM.NodeIndexes, kvmCluster.Status.Provider.NodeIndexes)
		if statusNeedsUpdate {
			kvmCluster.Status.Provider.NodeIndexes = cr.Status.KVM.NodeIndexes
			err = r.ctrlClient.Status().Update(ctx, &kvmCluster)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}
