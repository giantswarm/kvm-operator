package ingress

import (
	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvmtpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func newEtcdIngress(customObject kvmtpr.CustomObject) *extensionsv1.Ingress {
	ingress := &extensionsv1.Ingress{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: EtcdID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
				"app":      MasterID,
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
										ServiceName: MasterID,
										ServicePort: intstr.FromInt(2379),
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
