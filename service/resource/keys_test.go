package resource

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
