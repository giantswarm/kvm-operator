package pvc

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

const (
	// EtcdPVSize is the size the persistent volume for etcd is configured with.
	EtcdPVSize = "15Gi"
)

func newEtcdPVCs(cr v1alpha2.KVMCluster) ([]*corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []*corev1.PersistentVolumeClaim

	for i, node := range cr.Spec.Cluster.Nodes {
		if node.Role != key.MasterID {
			continue
		}

		quantity, err := resource.ParseQuantity(EtcdPVSize)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		persistentVolumeClaim := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: key.EtcdPVCName(key.ClusterID(&cr), key.VMNumber(i)),
				Labels: map[string]string{
					"app":      key.MasterID,
					"cluster":  key.ClusterID(&cr),
					"customer": key.ClusterCustomer(&cr),
					"node":     node.ID,
				},
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": StorageClass,
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: quantity,
					},
				},
			},
		}

		persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
	}

	return persistentVolumeClaims, nil
}
