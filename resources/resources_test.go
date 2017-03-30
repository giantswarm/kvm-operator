package resources

import (
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/cluster"
	"github.com/giantswarm/clustertpr/customer"
	"github.com/giantswarm/clustertpr/kubernetes"
	"github.com/giantswarm/clustertpr/kubernetes/api"
	"github.com/giantswarm/clustertpr/kubernetes/kubelet"
	"github.com/giantswarm/clustertpr/node"
	"github.com/giantswarm/kvmtpr"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestResourceComputation(t *testing.T) {
	// Master: Service, Ingress2379, Ingress6443, Deployment
	// Flannel-client: Deployment
	// Worker: Deployment, Service
	expectedObjects := 7

	cluster := clustertpr.Cluster{
		Cluster: cluster.Cluster{
			ID: "test",
		},
		Customer: customer.Customer{
			ID: "test",
		},
		Kubernetes: kubernetes.Kubernetes{
			API: api.API{
				ClusterIPRange: "10.3.0.0/31",
				InsecurePort:   8080,
				SecurePort:     6443,
			},
			Kubelet: kubelet.Kubelet{
				Port: 4194,
			},
		},
		Masters: []node.Node{
			node.Node{
				Memory: "4096",
				CPUs:   1,
			},
		},
		Workers: []node.Node{
			node.Node{
				Memory: "4096",
				CPUs:   1,
			},
		},
	}

	customObject := kvmtpr.CustomObject{
		Spec: kvmtpr.Spec{
			Cluster: cluster,
		},
	}

	objects, err := ComputeResources(customObject)
	if err != nil {
		t.Fatalf("Error when computing cluster resources %v", err)
	}

	fmt.Printf("%#v\n", objects)

	if len(objects) != expectedObjects {
		t.Fatalf("Number of objects in expected output differed from received units: %d != %d", len(objects), expectedObjects)
	}
}

// TestResourcesDontHaveClusterIDAsPrefix tests that resources do not have the
// cluster id as a prefix. This is due to IDs being alphanumeric, and Kubernetes
// not allowing resource names to begin with an integer.
func TestResourcesDontHaveClusterIDAsPrefix(t *testing.T) {
	id := "test"

	cluster := clustertpr.Cluster{
		Cluster: cluster.Cluster{
			ID: id,
		},
		Customer: customer.Customer{
			ID: id,
		},
		Kubernetes: kubernetes.Kubernetes{
			API: api.API{
				SecurePort:   6443,
				InsecurePort: 8080,
			},
			Kubelet: kubelet.Kubelet{
				Port: 4194,
			},
		},
		Masters: []node.Node{
			node.Node{
				Memory: "4096",
				CPUs:   1,
			},
		},
		Workers: []node.Node{
			node.Node{
				Memory: "4096",
				CPUs:   1,
			},
		},
	}

	customObject := kvmtpr.CustomObject{
		Spec: kvmtpr.Spec{
			Cluster: cluster,
		},
	}

	resources, err := ComputeResources(customObject)
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

	cluster := clustertpr.Cluster{
		Cluster: cluster.Cluster{
			ID: clusterID,
		},
		Customer: customer.Customer{
			ID: customerID,
		},
		Kubernetes: kubernetes.Kubernetes{
			API: api.API{
				SecurePort:   6443,
				InsecurePort: 8080,
			},
			Kubelet: kubelet.Kubelet{
				Port: 4194,
			},
		},
		Masters: []node.Node{
			node.Node{
				Memory: "4096",
				CPUs:   1,
			},
		},
		Workers: []node.Node{
			node.Node{
				Memory: "4096",
				CPUs:   1,
			},
		},
	}

	customObject := kvmtpr.CustomObject{
		Spec: kvmtpr.Spec{
			Cluster: cluster,
		},
	}

	resources, err := ComputeResources(customObject)
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
