package controller

import (
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"

	apiv1 "k8s.io/client-go/pkg/api/v1"
)

type ConfigMap interface {
	ClusterObj
}

type configMap struct {
	ClusterConfig
}

func (w *configMap) Create() error {
	configMap := &apiv1.ConfigMap{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "configmap",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "configmap",
			Labels: map[string]string{
				"cluster-id": w.ClusterID,
			},
		},
		Data: map[string]string{
			"vault-addr":                "http://mgmt.local.giantswarm.io:8200",
			"vault-token":               "a6fff246-a5da-4d03-aab0-423b44c0a6d7",
			"registry":                  "",
			"cluster-id":                "test",
			"customer-id":               "test",
			"cluster-vni":               "2",
			"cluster-backend":           "vxlan",
			"cluster-network":           "10.2.0.0/16",
			"etcd-port":                 "2379",
			"network-interface":         "ens33",
			"host-subnet-range":         "10.0.3.0/24",
			"bridge-network":            "brtest",
			"machine-mem":               "512",
			"machine-cpu-cores":         "4",
			"calico-subnet":             "192.168.0.0",
			"calico-cidr":               "16",
			"k8s-calico-mtu":            "1430",
			"k8s-cluster-ip-range":      "172.29.0.0",
			"k8s-cluster-ip-subnet":     "16",
			"k8s-version":               "1.4.0",
			"k8s-dns-ip":                "172.29.0.10",
			"k8s-domain":                "test.giantswarm.fra-1",
			"k8s-network-setup-version": "0.1",
			"k8s-etcd-prefix":           "test",
			"k8s-secure-port":           "6443",
			"k8s-insecure-port":         "8080",
			"k8s-etcd-domain-name":      "etcd.test.g8s.fra-1.giantswarm.io",
			"k8s-master-domain-name":    "api.test.g8s.fra-1.giantswarm.io",
			"k8s-master-service-name":   "k8s-master-vm",
			"k8s-master-port":           "6443",
			"k8s-node-labels":           "",
			"docker-extra-args":         "",
			"g8s-domain":                "g8s.giantswarm.fra-1",
			"g8s-dns-ip":                "172.31.0.10",
			"g8s-api-ip":                "172.31.0.1",
			"k8s-api-alt-names":         "kubernetes,kubernetes.default,kubernetes.default.svc,kubernetes.default.svc.cluster.local",
			"kemp-vs-ip":                "178.162.217.237",
			"kemp-vs-name":              "ingress-controller-test-k8s-gigantic-io",
			"kemp-vs-ports":             "80",
			"kemp-vs-ssl-acceleration":  "false",
			"kemp-user":                 "ddmmadmin",
			"kemp-password":             "xxxxx",
			"kemp-endpoint":             "https://178.162.217.249:8443/access/",
			"kemp-rs-port":              "30011",
			"kemp-vs-check-port":        "30010",
			"cloudflare-ip":             "178.162.217.237",
			"cloudflare-domain":         "gigantic.io",
			"cloudflare-token":          "xxxxxxx",
			"cloudflare-email":          "accounts@giantswarm.io",
		},
	}

	_, err := w.KubernetesClient.Core().ConfigMaps(w.Namespace).Create(configMap)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f *configMap)  Delete() error {
	return nil
}
