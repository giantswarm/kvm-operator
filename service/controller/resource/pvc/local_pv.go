package pvc

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/kvm-operator/service/controller/key"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note: Local volumes can only be used as a statically created PersistentVolume.
//       Dynamic provisioning is not supported.
func newLocalPVs(customObject v1alpha1.KVMConfig) ([]*corev1.PersistentVolume, error) {
	var persistentVolumes []*corev1.PersistentVolume

	filesystemVolumeMode := corev1.PersistentVolumeFilesystem

	for _, workerNode := range customObject.Spec.KVM.Workers {
		for _, hostVolume := range workerNode.HostVolumes {
			var persistentVolume = &corev1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: key.LocalPVCName(key.ClusterID(customObject), hostVolume.MountTag),
					Labels: map[string]string{
						"app":      key.WorkerID,
						"cluster":  key.ClusterID(customObject),
						"customer": key.ClusterCustomer(customObject),
						"node":     "",
					},
				},
				Spec: corev1.PersistentVolumeSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						Local: &corev1.LocalVolumeSource{
							Path: hostVolume.HostPath,
						},
					},
					StorageClassName:              LocalStorageClass,
					PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
					VolumeMode:                    &filesystemVolumeMode,
					Capacity: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: "",
					},
					NodeAffinity: &corev1.VolumeNodeAffinity{
						Required: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{},
										},
									},
								},
							},
						},
					},
				},
			}

			persistentVolumes = append(persistentVolumes, persistentVolume)
		}
	}


	return persistentVolumes, nil
}