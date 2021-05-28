package pvc

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

const (
	LabelMountTag = "mount-tag"
)

func newLocalPVCs(customObject v1alpha1.KVMConfig, persistentVolumes []corev1.PersistentVolume) ([]corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []corev1.PersistentVolumeClaim

	for i, workerKVM := range customObject.Spec.KVM.Workers {
		for _, hostVolume := range workerKVM.HostVolumes {
			pv, err := findPVByMountTag(persistentVolumes, hostVolume.MountTag)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			// discard the PV if is already bound to an existing PV
			if pv.Spec.ClaimRef != nil {
				return nil, microerror.Maskf(isAlreadyBound, "persistent volume %s is already bound to %s", pv.Name, pv.Spec.ClaimRef.Name)
			}

			persistentVolumeClaim := corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.LocalWorkerPVCName(key.ClusterID(customObject), key.VMNumber(i), hostVolume.MountTag),
					Labels: map[string]string{
						key.LabelCustomer:      key.ClusterCustomer(customObject),
						key.LabelApp:           key.WorkerID,
						key.LabelCluster:       key.ClusterID(customObject),
						key.LabelVersionBundle: key.OperatorVersion(customObject),
						"node":                 customObject.Spec.Cluster.Workers[i].ID,
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
							LabelMountTag: hostVolume.MountTag,
						},
					},
					StorageClassName: to.StringP(LocalStorageClass),
				},
			}

			persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
		}
	}

	return persistentVolumeClaims, nil
}

func findPVByMountTag(persistentVolumes []corev1.PersistentVolume, mountTag string) (corev1.PersistentVolume, error) {
	for _, persistentVolume := range persistentVolumes {
		if persistentVolume.ObjectMeta.Labels[LabelMountTag] != mountTag {
			continue
		}

		return persistentVolume, nil
	}

	return corev1.PersistentVolume{}, microerror.Maskf(notFoundError, "mount tag %s not found", mountTag)
}
