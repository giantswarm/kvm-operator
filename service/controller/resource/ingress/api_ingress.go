package ingress

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newAPIIngress(cr v1alpha2.KVMCluster) *networkingv1beta1.Ingress {
	ingress := &networkingv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking/v1beta",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: APIID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(cr),
				"customer": key.ClusterCustomer(cr),
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
						cr.Spec.Cluster.Kubernetes.API.Domain,
					},
				},
			},
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: cr.Spec.Cluster.Kubernetes.API.Domain,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: key.MasterID,
										ServicePort: intstr.FromInt(cr.Spec.Cluster.Kubernetes.API.SecurePort),
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
