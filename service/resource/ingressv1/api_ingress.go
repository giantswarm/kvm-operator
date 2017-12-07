package ingressv1

import (
	"github.com/giantswarm/kvmtpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv1"
)

func newAPIIngress(customObject kvmtpr.CustomObject) *extensionsv1.Ingress {
	ingress := &extensionsv1.Ingress{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: APIID,
			Labels: map[string]string{
				"cluster":  keyv1.ClusterID(customObject),
				"customer": keyv1.ClusterCustomer(customObject),
				"app":      keyv1.MasterID,
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
										ServiceName: keyv1.MasterID,
										ServicePort: intstr.FromInt(customObject.Spec.Cluster.Kubernetes.API.SecurePort),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}
