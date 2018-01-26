package pvcv2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/keyv2"
)

const (
	// EtcdPVSize is the size the persistent volume for etcd is configured with.
	EtcdPVSize = "15Gi"
)

func newEtcdPVCs(customObject v1alpha1.KVMConfig) ([]*apiv1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []*apiv1.PersistentVolumeClaim

	for i, masterNode := range customObject.Spec.Cluster.Masters {
		quantity, err := resource.ParseQuantity(EtcdPVSize)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		persistentVolumeClaim := &apiv1.PersistentVolumeClaim{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: keyv2.EtcdPVCName(keyv2.ClusterID(customObject), keyv2.VMNumber(i)),
				Labels: map[string]string{
					"app":      keyv2.MasterID,
					"cluster":  keyv2.ClusterID(customObject),
					"customer": keyv2.ClusterCustomer(customObject),
					"node":     masterNode.ID,
				},
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": StorageClass,
				},
			},
			Spec: apiv1.PersistentVolumeClaimSpec{
				AccessModes: []apiv1.PersistentVolumeAccessMode{
					apiv1.ReadWriteOnce,
				},
				Resources: apiv1.ResourceRequirements{
					Requests: map[apiv1.ResourceName]resource.Quantity{
						apiv1.ResourceStorage: quantity,
					},
				},
			},
		}

		persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
	}

	return persistentVolumeClaims, nil
}
