package service

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newMasterService(cr v1alpha2.KVMCluster) *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.MasterID,
			Namespace: key.ClusterID(cr),
			Labels: map[string]string{
				key.LegacyLabelCluster: key.ClusterID(cr),
				key.LabelCustomer:      key.ClusterCustomer(cr),
				key.LabelApp:           key.MasterID,
				key.LabelCluster:       key.ClusterID(cr),
				key.LabelOrganization:  key.ClusterCustomer(cr),
				key.LabelVersionBundle: key.OperatorVersion(cr),
			},
			Annotations: map[string]string{
				key.AnnotationEtcdDomain:        key.ClusterEtcdDomain(cr),
				key.AnnotationPrometheusCluster: key.ClusterID(cr),
				"prometheus.io/path":            "/healthz",
				"prometheus.io/port":            "30010",
				"prometheus.io/scheme":          "http",
				"prometheus.io/scrape":          "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "etcd",
					Port:     int32(2379),
					Protocol: "TCP",
				},
				{
					Name:     "api",
					Port:     int32(cr.Spec.Cluster.Kubernetes.API.SecurePort),
					Protocol: "TCP",
				},
			},
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	return service
}
