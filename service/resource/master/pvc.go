package master

import (
	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/resource"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (s *Service) newPersistentVolumeClaims(obj interface{}) ([]*apiv1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []*apiv1.PersistentVolumeClaim

	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	for i, masterNode := range customObject.Spec.Cluster.Masters {
		quantity, err := resource.ParseQuantity("15Gi")
		if err != nil {
			return nil, microerror.Maskf(err, "cant parse quantity")
		}

		persistentVolumeClaim := &apiv1.PersistentVolumeClaim{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: key.PVCName(key.ClusterID(*customObject), key.VMNumber(i)),
				Labels: map[string]string{
					"cluster":  key.ClusterID(*customObject),
					"customer": key.ClusterCustomer(*customObject),
					"app":      "master",
					"node":     masterNode.ID,
				},
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": "",
				},
			},
			Spec: apiv1.PersistentVolumeClaimSpec{
				AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
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
