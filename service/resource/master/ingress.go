package master

import (
	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func (s *Service) newIngresses(obj interface{}) ([]*extensionsv1.Ingress, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	ingresses := []*extensionsv1.Ingress{
		{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "extensions/v1beta",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: "etcd",
				Labels: map[string]string{
					"cluster":  key.ClusterID(*customObject),
					"customer": key.ClusterCustomer(*customObject),
					"app":      "master",
				},
				Annotations: map[string]string{
					"ingress.kubernetes.io/ssl-passthrough": "true",
				},
			},
			Spec: extensionsv1.IngressSpec{
				TLS: []extensionsv1.IngressTLS{
					{
						Hosts: []string{
							customObject.Spec.Cluster.Etcd.Domain,
						},
					},
				},
				Rules: []extensionsv1.IngressRule{
					{
						Host: customObject.Spec.Cluster.Etcd.Domain,
						IngressRuleValue: extensionsv1.IngressRuleValue{
							HTTP: &extensionsv1.HTTPIngressRuleValue{
								Paths: []extensionsv1.HTTPIngressPath{
									{
										Path: "/",
										Backend: extensionsv1.IngressBackend{
											ServiceName: "master",
											ServicePort: intstr.FromInt(2379),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "extensions/v1beta",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: "api",
				Labels: map[string]string{
					"cluster":  key.ClusterID(*customObject),
					"customer": key.ClusterCustomer(*customObject),
					"app":      "master",
				},
				Annotations: map[string]string{
					"ingress.kubernetes.io/ssl-passthrough": "true",
				},
			},
			Spec: extensionsv1.IngressSpec{
				TLS: []extensionsv1.IngressTLS{
					{
						Hosts: []string{
							customObject.Spec.Cluster.Kubernetes.API.Domain,
						},
					},
				},
				Rules: []extensionsv1.IngressRule{
					{
						Host: customObject.Spec.Cluster.Kubernetes.API.Domain,
						IngressRuleValue: extensionsv1.IngressRuleValue{
							HTTP: &extensionsv1.HTTPIngressRuleValue{
								Paths: []extensionsv1.HTTPIngressPath{
									{
										Path: "/",
										Backend: extensionsv1.IngressBackend{
											ServiceName: "master",
											ServicePort: intstr.FromInt(customObject.Spec.Cluster.Kubernetes.API.SecurePort),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingresses, nil
}
