package cleanupendpointips

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Resource_CleanupEndpointIPs_removeFromEndpointAddressList(t *testing.T) {
	testCases := []struct {
		EndpointAddressList         []corev1.EndpointAddress
		IndexesToRemove             []int
		ExpectedEndpointAddressList []corev1.EndpointAddress
	}{
		{
			EndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.2",
				},
				{
					IP: "10.0.0.3",
				},
			},
			IndexesToRemove: []int{0},
			ExpectedEndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.2",
				},
				{
					IP: "10.0.0.3",
				},
			},
		},
		{
			EndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.2",
				},
				{
					IP: "10.0.0.3",
				},
			},
			IndexesToRemove: []int{1},
			ExpectedEndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.3",
				},
			},
		},
		{
			EndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.2",
				},
				{
					IP: "10.0.0.3",
				},
			},
			IndexesToRemove: []int{2},
			ExpectedEndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.2",
				},
			},
		},
		{
			EndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.2",
				},
				{
					IP: "10.0.0.3",
				},
			},
			IndexesToRemove: []int{0, 2},
			ExpectedEndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.2",
				},
			},
		},
		{
			EndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.1",
				},
				{
					IP: "10.0.0.2",
				},
				{
					IP: "10.0.0.3",
				},
				{
					IP: "10.0.0.4",
				},
			},
			IndexesToRemove: []int{0, 1, 2},
			ExpectedEndpointAddressList: []corev1.EndpointAddress{
				{
					IP: "10.0.0.4",
				},
			},
		},
	}

	for i, tc := range testCases {
		returnedEndpointAddressList := removeFromEndpointAddressList(tc.EndpointAddressList, tc.IndexesToRemove)
		if !reflect.DeepEqual(returnedEndpointAddressList, tc.ExpectedEndpointAddressList) {
			t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedEndpointAddressList, returnedEndpointAddressList)
		}
	}
}

func Test_Resource_CleanupEndpointIPs_removeDeadIPFromEndpoints(t *testing.T) {
	testCases := []struct {
		Endpoints               *corev1.Endpoints
		Nodes                   []corev1.Node
		Pods                    []corev1.Pod
		ExpectedDeletedEndpoint int
		ExpectedEndpoints       *corev1.Endpoints
	}{
		// case 0 no dead ip
		{
			Endpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.2.3.4",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.1.1.1",
							},
						},
					},
				},
			},
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			ExpectedDeletedEndpoint: 0,
			ExpectedEndpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
		},
		// case 1 - 1 dead ip
		{
			Endpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
							{
								IP: "1.1.1.4",
							},
						},
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.2.3.4",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.1.1.1",
							},
						},
					},
				},
			},
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			ExpectedDeletedEndpoint: 1,
			ExpectedEndpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
		},
		// case 2 - 2 dead ips
		{
			Endpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
							{
								IP: "1.1.1.4",
							},
						},
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.2.3.4",
							},
						},
					},
				},
			},
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			ExpectedDeletedEndpoint: 2,
			ExpectedEndpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
		},
		// case 3 - 1 dead CP Pod
		{
			Endpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
							{
								IP: "1.1.1.4",
							},
						},
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.2.3.4",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.1.1.1",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode3",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							corev1.NodeAddress{
								Type:    corev1.NodeInternalIP,
								Address: "1.1.1.4",
							},
						},
					},
				},
			},
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode1",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode2",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testNode3",
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
				},
			},
			ExpectedDeletedEndpoint: 1,
			ExpectedEndpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
							{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		deletedEndpoints, returnedEndpoints, err := removeDeadIPFromEndpoints(tc.Endpoints, tc.Nodes, tc.Pods)
		if err != nil {
			t.Fatalf("case %d expected %d deleted endpoint ips but got error %v#", i, tc.ExpectedDeletedEndpoint, err)
		}

		if deletedEndpoints != tc.ExpectedDeletedEndpoint {
			t.Fatalf("case %d expected %d deleted endpoint ips got %d", i, tc.ExpectedDeletedEndpoint, deletedEndpoints)
		}

		if !reflect.DeepEqual(tc.ExpectedEndpoints, returnedEndpoints) {
			t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedEndpoints, returnedEndpoints)
		}
	}
}
func Test_Resource_CleanupEndpointIPs_podsEqualNodes(t *testing.T) {
	testCases := []struct {
		Pods           []corev1.Pod
		Nodes          []corev1.Node
		ExpectedResult bool
	}{
		// case 1 pods equal nodes
		{
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
			},
			ExpectedResult: true,
		},
		// case 2 pods equal nodes
		{
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-2",
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-2",
					},
				},
			},
			ExpectedResult: true,
		},
		// case 3 pods not equal nodes - missing node
		{
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-2",
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
			},
			ExpectedResult: false,
		},
		// case 4 pods not equal nodes - old node in TC API
		{
			Pods: []corev1.Pod{
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-2",
					},
				},
			},
			Nodes: []corev1.Node{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-1",
					},
				},
				{

					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-xxxx-test-OLD",
					},
				},
			},
			ExpectedResult: false,
		},
	}

	for i, tc := range testCases {
		result := podsEqualNodes(tc.Pods, tc.Nodes)

		if result != tc.ExpectedResult {
			t.Fatalf("case %d expected %t result got %t", i, tc.ExpectedResult, result)
		}
	}
}
