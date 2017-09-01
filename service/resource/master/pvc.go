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

		var etcdPVClaimName string = "pvc-master-etcd-" + key.ClusterID(*customObject) + "-" + key.VMNumber(i)

		persistentVolumeClaim := apiv1.PersistentVolumeClaim{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: etcdPVClaimName,
				Labels: map[string]string{
					"cluster":  key.ClusterID(*customObject),
					"customer": key.ClusterCustomer(*customObject),
					"app":      "master",
					"node":     masterNode.ID,
				},
			},
			Spec: apiv1.PersistentVolumeClaimSpec{
				AccessModes: apiv1.GetAccessModesFromString("ReadWriteOnce"),
				Resources: apiv1.ResourceRequirements{
					Requests: map[apiv1.ResourceName]resource.Quantity{
						"Storage": "15Gi",
					},
				},
			},
		}

		persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
	}

}
