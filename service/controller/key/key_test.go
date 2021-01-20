package key

import (
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/pkg/label"
)

func Test_ClusterID(t *testing.T) {
	expectedID := "abc12"

	cr := v1alpha2.KVMCluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				LabelCluster: "abc12",
			},
		},
	}

	if ClusterID(&cr) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterID(&cr))
	}
}

func Test_ClusterCustomer(t *testing.T) {
	expectedID := "test-customer"

	cr := v1alpha2.KVMCluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Organization: "abc-def",
			},
		},
	}

	if ClusterCustomer(&cr) != expectedID {
		t.Fatalf("Expected customer ID %s but was %s", expectedID, ClusterCustomer(&cr))
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

func Test_NodeClusterIP(t *testing.T) {
	testCases := []struct {
		Node         corev1.Node
		ExpectedIP   string
		ErrorMatcher func(error) bool
	}{
		// Test 1, node has an internal IP address
		{
			Node: corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "some-address",
						},
					},
				},
			},
			ExpectedIP:   "some-address",
			ErrorMatcher: nil,
		},
		// Test 2, node doesn't have an internal IP address
		{
			Node:         corev1.Node{},
			ExpectedIP:   "",
			ErrorMatcher: IsMissingNodeInternalIP,
		},
	}

	for i, tc := range testCases {
		ip, err := NodeInternalIP(tc.Node)

		switch {
		case err == nil && tc.ErrorMatcher == nil:
			// correct; carry on
		case err != nil && tc.ErrorMatcher == nil:
			t.Fatalf("error == %#v, want nil", err)
		case err == nil && tc.ErrorMatcher != nil:
			t.Fatalf("error == nil, want non-nil")
		case !tc.ErrorMatcher(err):
			t.Fatalf("error == %#v, want matching", err)
		case tc.ErrorMatcher(err):
			return
		}

		if ip != tc.ExpectedIP {
			t.Fatalf("case %d expected %T got %T", i+1, tc.ExpectedIP, ip)
		}
	}
}

func Test_PortMappings(t *testing.T) {
	cr := v1alpha2.KVMCluster{
		Spec: v1alpha2.KVMClusterSpec{
			Provider: v1alpha2.KVMClusterSpecProvider{
				PortMappings: []v1alpha2.KVMClusterSpecProviderPortMappings{
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

	actual := PortMappings(cr)

	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Expected port mapping %+v but was %+v", expected[i], actual[i])
		}
	}
}

func Test_PortMappings_CompatibilityMode(t *testing.T) {
	cr := v1alpha2.KVMCluster{
		Spec: v1alpha2.KVMClusterSpec{
			Provider: v1alpha2.KVMClusterSpecProvider{
				PortMappings: []v1alpha2.KVMClusterSpecProviderPortMappings{},
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

	actual := PortMappings(cr)

	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Expected port mapping %+v but was %+v", expected[i], actual[i])
		}
	}
}
