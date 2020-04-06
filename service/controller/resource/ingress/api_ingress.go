package ingress

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newAPIIngress(customObject v1alpha1.KVMConfig) *networkingv1beta1.Ingress {
	ingress := &networkingv1beta1.Ingress{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: APIID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
				"app":      key.MasterID,
			},
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough": "true",
			},
		},
		Spec: networkingv1beta1.IngressSpec{
			TLS: []networkingv1beta1.IngressTLS{
				{
					Hosts: []string{
						customObject.Spec.Cluster.Kubernetes.API.Domain,
					},
				},
			},
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: customObject.Spec.Cluster.Kubernetes.API.Domain,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: key.MasterID,
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
