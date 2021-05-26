package pvc

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

const (
	LabelMountTag = "mount-tag"
)

func newLocalPVCs(customObject v1alpha1.KVMConfig, pvsList *corev1.PersistentVolumeList) ([]*corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []*corev1.PersistentVolumeClaim
	localStorageClass := LocalStorageClass

	for i, workerKVM := range customObject.Spec.KVM.Workers {
		for _, hostVolume := range workerKVM.HostVolumes {

			pv := findPV(pvsList, hostVolume.MountTag)
			if pv == nil {
				return nil, microerror.Maskf(notFoundError, "mount tag %s is not available", hostVolume.MountTag)
			}

			persistentVolumeClaim := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.LocalWorkerPVCName(key.ClusterID(customObject), key.VMNumber(i)),
					Labels: map[string]string{
						key.LegacyLabelCluster: key.ClusterID(customObject),
						key.LabelCustomer:      key.ClusterCustomer(customObject),
						key.LabelApp:           key.WorkerID,
						key.LabelCluster:       key.ClusterID(customObject),
						key.LabelOrganization:  key.ClusterCustomer(customObject),
						key.LabelVersionBundle: key.OperatorVersion(customObject),
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
					StorageClassName: &localStorageClass,
				},
			}

			persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
		}
	}

	return persistentVolumeClaims, nil
}

func findPV(pvsList *corev1.PersistentVolumeList, mountTag string) *corev1.PersistentVolume {
	for _, pv := range pvsList.Items {
		if pv.ObjectMeta.Labels[LabelMountTag] != mountTag {
			continue
		}

		return &pv
	}

	return nil
}
