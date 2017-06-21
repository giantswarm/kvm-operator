package resource

import (
	"net"
	"testing"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/cluster"
	"github.com/giantswarm/clustertpr/customer"
	"github.com/giantswarm/kvmtpr"
)

func TestClusterID(t *testing.T) {
	expectedID := "test-cluster"

	cluster := clustertpr.Cluster{
		Cluster: cluster.Cluster{
			ID: expectedID,
		},
		Customer: customer.Customer{
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

func TestClusterCustomer(t *testing.T) {
	expectedID := "test-customer"

	cluster := clustertpr.Cluster{
		Cluster: cluster.Cluster{
			ID: "test-cluster",
		},
		Customer: customer.Customer{
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
