package service

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v14/key"
)

func newMasterService(customObject v1alpha1.KVMConfig) *apiv1.Service {
	service := &apiv1.Service{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.MasterID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
				"app":      key.MasterID,
				"giantswarm.io/cluster":        key.ClusterID(customObject),
				"giantswarm.io/etcd_domain":    key.ClusterEtcdDomain(customObject),
				"giantswarm.io/organization":   key.ClusterCustomer(customObject),
				"giantswarm.io/version_bundle": key.VersionBundleVersion(customObject),
			},
			Annotations: map[string]string{
				"giantswarm.io/prometheus-cluster": key.ClusterID(customObject),
				"prometheus.io/path":               "/healthz",
				"prometheus.io/port":               "30010",
				"prometheus.io/scheme":             "http",
				"prometheus.io/scrape":             "true",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{
				{
					Name:     "etcd",
					Port:     int32(2379),
					Protocol: "TCP",
				},
				{
					Name:     "api",
					Port:     int32(customObject.Spec.Cluster.Kubernetes.API.SecurePort),
					Protocol: "TCP",
				},
			},
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	return service
}
