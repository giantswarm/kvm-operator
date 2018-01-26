package ingressv2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/keyv2"
)

func newAPIIngress(customObject v1alpha1.KVMConfig) *extensionsv1.Ingress {
	ingress := &extensionsv1.Ingress{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: APIID,
			Labels: map[string]string{
				"cluster":  keyv2.ClusterID(customObject),
				"customer": keyv2.ClusterCustomer(customObject),
				"app":      keyv2.MasterID,
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
										ServiceName: keyv2.MasterID,
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
