package resources

import (
	"log"
	"testing"

	"github.com/giantswarm/clusterspec"
)

func TestComponent_Creation(t *testing.T) {
	// Master: Service, Ingress2379, Ingress6443, Deployment
	// Flannel-client: Deployment
	// Worker: Deployment, Service
	// Ingress controller: Deployment, Service
	expectedObjects := 9

	cluster := &clusterspec.Cluster{
		Spec: clusterspec.ClusterSpec{
			Customer:  "test",
			ClusterId: "test",
		},
	}

	cluster.Spec.Worker.Replicas = int32(1)
	cluster.Spec.Worker.WorkerServicePort = "4194"
	cluster.Spec.Master.SecurePort = "6443"

	objects, err := ComputeResources(cluster)
	if err != nil {
		t.Fatalf("Error when computing cluster resources %v", err)
	}

	for _, obj := range objects {
		log.Println("obj desired resources for cluster: %v", obj)
	}

	if len(objects) != expectedObjects {
		t.Fatalf("Number of objects in expected output differed from received units: %d != %d", len(objects), expectedObjects)
	}
}
