package pvc

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/v4/pkg/label"
	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

const (
	// EtcdPVSize is the size the persistent volume for etcd is configured with.
	EtcdPVSize = "15Gi"
)

func (r *Resource) getDesiredMasterPVCs(customObject v1alpha1.KVMConfig) ([]corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []corev1.PersistentVolumeClaim

	for i, masterNode := range customObject.Spec.Cluster.Masters {
		quantity, err := resource.ParseQuantity(EtcdPVSize)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		persistentVolumeClaim := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: key.EtcdPVCName(key.ClusterID(customObject), key.VMNumber(i)),
				Labels: map[string]string{
					label.ManagedBy:        key.OperatorName,
					key.LabelCustomer:      key.ClusterCustomer(customObject),
					key.LabelApp:           key.MasterID,
					key.LabelCluster:       key.ClusterID(customObject),
					key.LegacyLabelCluster: key.ClusterID(customObject),
					key.LabelVersionBundle: key.OperatorVersion(customObject),
					"node":                 masterNode.ID,
				},
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": EtcdStorageClass,
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
