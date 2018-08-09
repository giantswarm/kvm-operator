package key

import (
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_ClusterID(t *testing.T) {
	expectedID := "test-cluster"

	customObject := v1alpha1.KVMConfig{
		Spec: v1alpha1.KVMConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: expectedID,
				Customer: v1alpha1.ClusterCustomer{
					ID: "test-customer",
				},
			},
		},
	}

	if ClusterID(customObject) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterID(customObject))
	}
}

func Test_ClusterCustomer(t *testing.T) {
	expectedID := "test-customer"

	customObject := v1alpha1.KVMConfig{
		Spec: v1alpha1.KVMConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: expectedID,
				Customer: v1alpha1.ClusterCustomer{
					ID: "test-customer",
				},
			},
		},
	}

	if ClusterCustomer(customObject) != expectedID {
		t.Fatalf("Expected customer ID %s but was %s", expectedID, ClusterCustomer(customObject))
	}
}

func Test_NetworkDNSBlock(t *testing.T) {
	dnsServers := NetworkDNSBlock([]net.IP{
		net.ParseIP("8.8.8.8"),
		net.ParseIP("8.8.4.4"),
	})

	expected := `DNS=8.8.8.8
DNS=8.8.4.4`

	if dnsServers != expected {
		t.Fatal("expected", expected, "got", dnsServers)
	}
}

func Test_PortMappings(t *testing.T) {
	customObject := v1alpha1.KVMConfig{
		Spec: v1alpha1.KVMConfigSpec{
			KVM: v1alpha1.KVMConfigSpecKVM{
				PortMappings: []v1alpha1.KVMConfigSpecKVMPortMappings{
					{
						Name:       "ingress-http",
						NodePort:   31010,
						TargetPort: 30010,
					},
					{
						Name:       "ingress-https",
						NodePort:   31011,
						TargetPort: 30011,
					},
				},
			},
		},
	}

	expected := []corev1.ServicePort{
		{
			Name:       "ingress-http",
			NodePort:   int32(31010),
			Port:       int32(30010),
			TargetPort: intstr.FromInt(30010),
		},
		{
			Name:       "ingress-https",
			NodePort:   int32(31011),
			Port:       int32(30011),
			TargetPort: intstr.FromInt(30011),
		},
	}

	actual := PortMappings(customObject)

	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Expected port mapping %+v but was %+v", expected[i], actual[i])
		}
	}
}

func Test_PortMappings_CompatibilityMode(t *testing.T) {
	customObject := v1alpha1.KVMConfig{
		Spec: v1alpha1.KVMConfigSpec{
			KVM: v1alpha1.KVMConfigSpecKVM{
				PortMappings: []v1alpha1.KVMConfigSpecKVMPortMappings{},
			},
		},
	}

	expected := []corev1.ServicePort{
		{
			Name:       "http",
			Port:       int32(30010),
			TargetPort: intstr.FromInt(30010),
		},
		{
			Name:       "https",
			Port:       int32(30011),
			TargetPort: intstr.FromInt(30011),
		},
	}

	actual := PortMappings(customObject)

	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Expected port mapping %+v but was %+v", expected[i], actual[i])
		}
	}
}
