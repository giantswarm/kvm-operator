package servicev2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func newWorkerService(customObject v1alpha1.KvmConfig) *apiv1.Service {
	service := &apiv1.Service{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: keyv2.WorkerID,
			Labels: map[string]string{
				"cluster":  keyv2.ClusterID(customObject),
				"customer": keyv2.ClusterCustomer(customObject),
				"app":      keyv2.WorkerID,
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{
				{
					Name:       "http",
					Port:       int32(30010),
					Protocol:   apiv1.ProtocolTCP,
					TargetPort: intstr.FromInt(30010),
				},
				{
					Name:       "https",
					Port:       int32(30011),
					Protocol:   apiv1.ProtocolTCP,
					TargetPort: intstr.FromInt(30011),
				},
			},
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	return service
}
