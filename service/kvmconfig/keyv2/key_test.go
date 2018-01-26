package keyv2

import (
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
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

func Test_HasNodeController(t *testing.T) {
	testCases := []struct {
		Obj            v1alpha1.KVMConfig
		ExpectedResult bool
	}{
		{
			Obj: v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					KVM: v1alpha1.KVMConfigSpecKVM{
						K8sKVM: v1alpha1.KVMConfigSpecKVMK8sKVM{
							Docker: v1alpha1.KVMConfigSpecKVMK8sKVMDocker{
								Image: "123",
							},
						},
						NodeController: v1alpha1.KVMConfigSpecKVMNodeController{
							Docker: v1alpha1.KVMConfigSpecKVMNodeControllerDocker{
								Image: "123",
							},
						},
					},
				},
			},
			ExpectedResult: true,
		},
		{
			Obj: v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					KVM: v1alpha1.KVMConfigSpecKVM{
						K8sKVM: v1alpha1.KVMConfigSpecKVMK8sKVM{
							Docker: v1alpha1.KVMConfigSpecKVMK8sKVMDocker{
								Image: "123",
							},
						},
					},
				},
			},
			ExpectedResult: false,
		},
	}

	for i, tc := range testCases {
		ActualResult := HasNodeController(tc.Obj)

		if ActualResult != tc.ExpectedResult {
			t.Fatalf("Case %d expected %t got %t", i+1, tc.ExpectedResult, ActualResult)
		}
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
