package worker

import (
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/resource"
)

func (s *Service) newService(obj interface{}) (*apiv1.Service, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	// TODO maybe add one service per worker?

	service := &apiv1.Service{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: "worker",
			Labels: map[string]string{
				"cluster":  resource.ClusterID(*customObject),
				"customer": resource.ClusterCustomer(*customObject),
				"app":      "worker",
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

	return service, nil
}
