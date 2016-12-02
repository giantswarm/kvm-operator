package resources

import (
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"

	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

type ConfigMap interface {
	ClusterObj
}

type configMap struct {
	Cluster
}

func (c *configMap) GenerateResources() ([]runtime.Object, error) {
	configMap := &apiv1.ConfigMap{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "configmap",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "configmap",
			Labels: map[string]string{
				"cluster-id": c.Spec.ClusterID,
			},
		},
		Data: map[string]string{

			"calico-subnet":             "192.168.0.0",
			"calico-cidr":               "16",
			"k8s-calico-mtu":            "1430",
			"k8s-cluster-ip-range":      "172.29.0.0",
			"k8s-cluster-ip-subnet":     "16",
			"k8s-network-setup-version": "0.1",

			"k8s-master-service-name":   "",
			"k8s-master-port":           "6443",
			"k8s-node-labels":           "",
			"docker-extra-args":         "",
			"g8s-domain":                "g8s.giantswarm.fra-1",
			"g8s-dns-ip":                "172.31.0.10",
			"g8s-api-ip":                "172.31.0.1",

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

	objects := append([]runtime.Object{}, configMap)

	return objects, nil
}
