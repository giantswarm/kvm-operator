package ingress

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newEtcdIngress(cr v1alpha2.KVMCluster) *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking/v1beta",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: EtcdID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(cr),
				"customer": key.ClusterCustomer(cr),
				"app":      key.MasterID,
			},
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough": "true",
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts: []string{
						cr.Spec.Cluster.Etcd.Domain,
					},
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: cr.Spec.Cluster.Etcd.Domain,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: key.MasterID,
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
