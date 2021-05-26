package pvc

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newLocalPVCs(customObject v1alpha1.KVMConfig, pvsList *corev1.PersistentVolumeList) ([]*corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []*corev1.PersistentVolumeClaim

	for i, workerKVM := range customObject.Spec.KVM.Workers {
		for _, hostVolume := range workerKVM.HostVolumes {
			for _, pv := range pvsList.Items {
				if pv.ObjectMeta.Labels["mount-tag"] == hostVolume.MountTag {
					persistentVolumeClaim := &corev1.PersistentVolumeClaim{
						TypeMeta: metav1.TypeMeta{
							Kind:       "PersistentVolumeClaim",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: key.LocalWorkerPVCName(key.ClusterID(customObject), key.VMNumber(i)),
							Labels: map[string]string{
								"app":      key.WorkerID,
								"cluster":  key.ClusterID(customObject),
								"customer": key.ClusterCustomer(customObject),
								"node":     "",
							},
							Annotations: map[string]string{
								"volume.beta.kubernetes.io/storage-class": LocalStorageClass,
							},
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							Resources: corev1.ResourceRequirements{
								Requests: pv.Spec.Capacity,
							},
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"mount-tag": hostVolume.MountTag,
								},
							},
						},
					}

					persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
				}
			}
		}
	}

	return persistentVolumeClaims, nil
}
