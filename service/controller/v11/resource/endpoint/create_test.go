package endpoint

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoint_ApplyCreateChange(t *testing.T) {
	testCases := []struct {
		CreateState       *corev1.Endpoints
		ExpectedEndpoints []*corev1.Endpoints
	}{
		{
			CreateState: &corev1.Endpoints{
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
								{
									IP: "1.1.1.1",
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
		fakeK8sClient := fake.NewSimpleClientset()
		var newResource *Resource
		{
			resourceConfig := DefaultConfig()
			resourceConfig.K8sClient = fakeK8sClient
			resourceConfig.Logger = microloggertest.New()
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
		}
		err := newResource.ApplyCreateChange(resourcecanceledcontext.NewContext(context.TODO(), make(chan struct{})), nil, tc.CreateState)
		if err != nil {
			t.Fatal("case", i, "expected", nil, "got", err)
		}
		for _, k8sEndpoint := range tc.ExpectedEndpoints {
			returnedEndpoint, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Get(k8sEndpoint.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("%d: error returned setting up k8s endpoints: %s\n", i, err)
			}
			if !reflect.DeepEqual(k8sEndpoint, returnedEndpoint) {
				t.Fatalf("case %d expected %#v got %#v", i, k8sEndpoint, returnedEndpoint)
			}
		}
	}
}

func Test_Resource_Endpoint_newCreateChange(t *testing.T) {
	testCases := []struct {
		CurrentState        *Endpoint
		DesiredState        *Endpoint
		ExpectedCreateState *corev1.Endpoints
		Obj                 interface{}
		SetupService        *corev1.Service
	}{
		{
			CurrentState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
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
			SetupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 1234,
						},
					},
				},
			},
			ExpectedCreateState: nil,
		},
		{
			CurrentState: nil,
			DesiredState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			SetupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 1234,
						},
					},
				},
			},
			ExpectedCreateState: &corev1.Endpoints{
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
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
		},
		{
			CurrentState: nil,
			DesiredState: &Endpoint{
				IPs: []string{
					"5.5.5.5",
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			SetupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 1234,
						},
					},
				},
			},
			ExpectedCreateState: &corev1.Endpoints{
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
			CurrentState: nil,
			DesiredState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			SetupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 1234,
						},
					},
				},
			},
			ExpectedCreateState: &corev1.Endpoints{
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
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		var err error
		var newResource *Resource
		{
			resourceConfig := DefaultConfig()
			resourceConfig.K8sClient = fake.NewSimpleClientset()
			resourceConfig.Logger = microloggertest.New()
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
		}

		if tc.SetupService != nil {
			if _, err := newResource.k8sClient.CoreV1().Services(tc.SetupService.Namespace).Create(tc.SetupService); err != nil {
				t.Fatalf("%d: error returned setting up k8s service: %s\n", i, err)
			}
		}

		result, err := newResource.newCreateChange(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatal("case", i, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedCreateState, result) {
			t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedCreateState, result)
		}
	}
}
