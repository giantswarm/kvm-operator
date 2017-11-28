package key

import (
	"net"
	"testing"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	kvmtprspec "github.com/giantswarm/kvmtpr/spec"
	kvmtprspeckvm "github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/kvmtpr/spec/kvm/k8skvm"
	"github.com/giantswarm/kvmtpr/spec/kvm/nodecontroller"
)

func Test_ClusterID(t *testing.T) {
	expectedID := "test-cluster"

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: expectedID,
		},
		Customer: spec.Customer{
			ID: "test-customer",
		},
	}

	customObject := kvmtpr.CustomObject{
		Spec: kvmtpr.Spec{
			Cluster: cluster,
		},
	}

	if ClusterID(customObject) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterID(customObject))
	}
}

func Test_ClusterCustomer(t *testing.T) {
	expectedID := "test-customer"

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
		Customer: spec.Customer{
			ID: expectedID,
		},
	}

	customObject := kvmtpr.CustomObject{
		Spec: kvmtpr.Spec{
			Cluster: cluster,
		},
	}

	if ClusterCustomer(customObject) != expectedID {
		t.Fatalf("Expected customer ID %s but was %s", expectedID, ClusterCustomer(customObject))
	}
}

func Test_HasNodeController(t *testing.T) {
	testCases := []struct {
		Obj            kvmtpr.CustomObject
		ExpectedResult bool
	}{
		{
			Obj: kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							Docker: k8skvm.Docker{
								Image: "123",
							},
						},
						NodeController: kvmtprspeckvm.NodeController{
							Docker: nodecontroller.Docker{
								Image: "123",
							},
						},
					},
				},
			},
			ExpectedResult: true,
		},
		{
			Obj: kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							Docker: k8skvm.Docker{
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
