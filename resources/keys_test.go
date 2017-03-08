package resources

import (
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
			ID: "test-cluster",
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

func TestCustomerID(t *testing.T) {
	expectedID := "test-customer"

	cluster := clustertpr.Cluster{
		Cluster: cluster.Cluster{
			ID: "test-cluster",
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

	if CustomerID(customObject) != expectedID {
		t.Fatalf("Expected customer ID %s but was %s", expectedID, CustomerID(customObject))
	}
}
