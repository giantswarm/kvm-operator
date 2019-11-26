package service

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v22/key"
)

func newWorkerService(customObject v1alpha1.KVMConfig) *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.WorkerID,
			Namespace: key.ClusterID(customObject),
			Labels: map[string]string{
				key.LegacyLabelCluster: key.ClusterID(customObject),
				key.LabelCustomer:      key.ClusterCustomer(customObject),
				key.LabelApp:           key.MasterID,
				key.LabelCluster:       key.ClusterID(customObject),
				key.LabelOrganization:  key.ClusterCustomer(customObject),
				key.LabelVersionBundle: key.VersionBundleVersion(customObject),
			},
			Annotations: map[string]string{
				"prometheus.io/path":   "/healthz",
				"prometheus.io/port":   "30010",
				"prometheus.io/scheme": "http",
				"prometheus.io/scrape": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: key.PortMappings(customObject),
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	return service
}
