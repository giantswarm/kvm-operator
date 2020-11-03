package endpoint

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	g8sfake "github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoint_ApplyDeleteChange(t *testing.T) {
	testCases := []struct {
		DeleteState       *corev1.Endpoints
		ExpectedEndpoints []*corev1.Endpoints
		SetupEndpoints    []*corev1.Endpoints
	}{
		{
			DeleteState: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{},
					},
				},
			},
			ExpectedEndpoints: nil,
			SetupEndpoints: []*corev1.Endpoints{
				{
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
		},
		{
			DeleteState: &corev1.Endpoints{
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
			ExpectedEndpoints: []*corev1.Endpoints{
				{
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
			SetupEndpoints: []*corev1.Endpoints{
				{
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
		},
	}
	var err error

	for i, tc := range testCases {
		fakeG8sClient := g8sfake.NewSimpleClientset()
		fakeK8sClient := fake.NewSimpleClientset()

		var newResource *Resource
		{
			c := Config{
				G8sClient: fakeG8sClient,
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			newResource, err = New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
		}

		for _, k8sEndpoint := range tc.SetupEndpoints {
			if _, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(context.Background(), k8sEndpoint, metav1.CreateOptions{}); err != nil {
				t.Fatalf("%d: error returned setting up k8s endpoints: %s\n", i, err)
			}
		}

		err := newResource.ApplyDeleteChange(resourcecanceledcontext.NewContext(context.TODO(), make(chan struct{})), nil, tc.DeleteState)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		for _, k8sEndpoint := range tc.ExpectedEndpoints {
			returnedEndpoint, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Get(context.Background(), k8sEndpoint.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("%d: error returned setting up k8s endpoints: %s\n", i+1, err)
			}
			if !reflect.DeepEqual(k8sEndpoint, returnedEndpoint) {
				t.Fatalf("case %d expected %#v got %#v", i+1, k8sEndpoint, returnedEndpoint)
			}
		}
	}
}

func Test_Resource_Endpoint_ipsForDeleteChange(t *testing.T) {
	testCases := []struct {
		name        string
		base        []string
		cutset      []string
		expectedSet []string
	}{
		{
			name: "case 0: Base has one more address than the cutset",
			base: []string{
				"1.1.1.1",
				"0.0.0.0",
			},
			cutset: []string{
				"1.1.1.1",
			},
			expectedSet: []string{
				"0.0.0.0",
			},
		},
		{
			name: "case 1: Base and cutset have the same addresses",
			base: []string{
				"1.1.1.1",
			},
			cutset: []string{
				"1.1.1.1",
			},
			expectedSet: []string{},
		},
		{
			name: "case 2: Cutset has one more address than base",
			base: []string{
				"1.1.1.1",
			},
			cutset: []string{
				"1.1.1.1",
				"0.0.0.0",
			},
			expectedSet: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			resultSet := ipsForDeleteChange(tc.base, tc.cutset)

			if !reflect.DeepEqual(resultSet, tc.expectedSet) {
				t.Fatalf("resultSet == %q, want %q", resultSet, tc.expectedSet)
			}
		})
	}
}

func Test_Resource_Endpoint_newDeleteChange(t *testing.T) {
	testCases := []struct {
		CurrentState        *Endpoint
		DesiredState        *Endpoint
		ExpectedDeleteState *corev1.Endpoints
		Obj                 interface{}
	}{
		{
			CurrentState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				Ports: []corev1.EndpointPort{
					{
						Port: 1234,
					},
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			Obj: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
				},
			},
			ExpectedDeleteState: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Ports: []corev1.EndpointPort{
							{
								Port: 1234,
							},
						},
					},
				},
			},
		},
		{
			CurrentState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
					"1.2.3.4",
				},
				Ports: []corev1.EndpointPort{
					{
						Port: 1234,
					},
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			Obj: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
				},
			},
			ExpectedDeleteState: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Ports: []corev1.EndpointPort{
							{
								Port: 1234,
							},
						},
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
		},
		{
			CurrentState: &Endpoint{
				IPs: []string{
					"5.5.5.5",
					"1.2.3.4",
				},
				Ports: []corev1.EndpointPort{
					{
						Port: 1234,
					},
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			Obj: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
				},
			},
			ExpectedDeleteState: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Ports: []corev1.EndpointPort{
							{
								Port: 1234,
							},
						},
						Addresses: []corev1.EndpointAddress{
							{
								IP: "5.5.5.5",
							},
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
		},
		{
			CurrentState: &Endpoint{
				Ports: []corev1.EndpointPort{
					{
						Port: 1234,
					},
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			DesiredState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			Obj: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
				},
			},
			ExpectedDeleteState: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Ports: []corev1.EndpointPort{
							{
								Port: 1234,
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var newResource *Resource
			{
				c := Config{
					G8sClient: g8sfake.NewSimpleClientset(),
					K8sClient: fake.NewSimpleClientset(),
					Logger:    microloggertest.New(),
				}

				newResource, err = New(c)
				if err != nil {
					t.Fatal("expected", nil, "got", err)
				}
			}

			result, err := newResource.newDeleteChange(context.Background(), tc.Obj, tc.CurrentState, tc.DesiredState)
			if err != nil {
				t.Fatal("case", i, "expected", nil, "got", err)
			}
			if !reflect.DeepEqual(tc.ExpectedDeleteState, result) {
				t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedDeleteState, result)
			}
		})
	}
}
