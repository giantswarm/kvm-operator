package resources

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/clusterspec"
)

func TestResourceComputation(t *testing.T) {
	// Master: Service, Ingress2379, Ingress6443, Deployment
	// Flannel-client: Deployment
	// Worker: Deployment, Service
	expectedObjects := 7

	cluster := clusterspec.Cluster{
		Spec: clusterspec.ClusterSpec{
			ClusterId: "test",
			Customer:  "test",
		},
	}

	cluster.Spec.Worker.Replicas = int32(1)
	cluster.Spec.Worker.WorkerServicePort = "4194"
	cluster.Spec.Master.SecurePort = "6443"
	cluster.Spec.Master.InsecurePort = "8080"

	objects, err := ComputeResources(cluster)
	if err != nil {
		t.Fatalf("Error when computing cluster resources %v", err)
	}

	if len(objects) != expectedObjects {
		t.Fatalf("Number of objects in expected output differed from received units: %d != %d", len(objects), expectedObjects)
	}
}

// TestResourcesDontHaveClusterIDAsPrefix tests that resources do not have the
// cluster id as a prefix. This is due to IDs being alphanumeric, and Kubernetes
// not allowing resource names to begin with an integer.
func TestResourcesDontHaveClusterIDAsPrefix(t *testing.T) {
	id := "test"

	cluster := clusterspec.Cluster{
		Spec: clusterspec.ClusterSpec{
			ClusterId: id,
			Customer:  id,
		},
	}

	cluster.Spec.Worker.Replicas = int32(1)
	cluster.Spec.Worker.WorkerServicePort = "4194"
	cluster.Spec.Master.SecurePort = "6443"
	cluster.Spec.Master.InsecurePort = "8080"

	resources, err := ComputeResources(cluster)
	if err != nil {
		t.Fatalf("Error when computing cluster resources: %v", err)
	}

	failIfPrefixFound := func(prefix, resourceName string) {
		if strings.HasPrefix(resourceName, prefix) {
			t.Fatalf(fmt.Sprintf("Prefix %v found for resource: %v\n", prefix, resourceName))
		}
	}

	for _, resource := range resources {
		switch r := resource.(type) {
		case *v1.ConfigMap:
			failIfPrefixFound(id, r.Name)
		case *v1.Service:
			failIfPrefixFound(id, r.Name)
		case *v1beta1.Deployment:
			failIfPrefixFound(id, r.Name)
			failIfPrefixFound(id, r.Spec.Template.Name)
			failIfPrefixFound(id, r.Spec.Template.GenerateName)
		case *v1beta1.Ingress:
			failIfPrefixFound(id, r.Name)
		default:
			t.Fatalf("Could not determine resource type\n")
		}
	}
}

func TestResourcesHaveCorrectLabelScheme(t *testing.T) {
	clusterID := "cluster-test"
	customerID := "customer-test"

	cluster := clusterspec.Cluster{
		Spec: clusterspec.ClusterSpec{
			ClusterId: clusterID,
			Customer:  customerID,
		},
	}

	cluster.Spec.Worker.Replicas = int32(1)
	cluster.Spec.Worker.WorkerServicePort = "4194"
	cluster.Spec.Master.SecurePort = "6443"
	cluster.Spec.Master.InsecurePort = "8080"

	resources, err := ComputeResources(cluster)
	if err != nil {
		t.Fatalf("Error when computing cluster resources: %v", err)
	}

	failIfWrongLabelScheme := func(name string, labels map[string]string) {
		cluster, ok := labels["cluster"]
		if !ok {
			t.Fatalf(fmt.Sprintf("Could not find 'cluster' label for resource: %v\n", name))
		}
		if cluster != clusterID {
			t.Fatalf(fmt.Sprintf("Resource did not have correct 'cluster' label: %v, %v", name, cluster))
		}

		customer, ok := labels["customer"]
		if !ok {
			t.Fatalf(fmt.Sprintf("Could not find 'customer' label for resource: %v\n", name))
		}
		if customer != customerID {
			t.Fatalf(fmt.Sprintf("Resource did not have correct 'customer' label: %v, %v", name, customer))
		}

		app, ok := labels["app"]
		if !ok {
			t.Fatalf(fmt.Sprintf("Could not find 'app' label for resource: %v\n", name))
		}
		// App will be different for each kind of resource (master, worker, etc.)
		// Checking existence is good enough
		if app == "" {
			t.Fatalf(fmt.Sprintf("Resourcer had empty 'app' label: %v", name))
		}
	}

	for _, resource := range resources {
		switch r := resource.(type) {
		case *v1.ConfigMap:
			failIfWrongLabelScheme(r.Name, r.Labels)
		case *v1.Service:
			failIfWrongLabelScheme(r.Name, r.Labels)
		case *v1beta1.Deployment:
			failIfWrongLabelScheme(r.Name, r.Labels)
			failIfWrongLabelScheme(r.Name, r.Spec.Template.Labels)
		case *v1beta1.Ingress:
			failIfWrongLabelScheme(r.Name, r.Labels)
		default:
			t.Fatalf("Could not determine resource type\n")
		}
	}
}
