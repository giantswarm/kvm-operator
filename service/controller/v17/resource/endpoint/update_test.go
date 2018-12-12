package endpoint

import (
	"context"
	"reflect"
	"testing"

	g8sfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoint_newUpdateChange(t *testing.T) {
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
			CurrentState: &Endpoint{
				IPs: []string{
					"1.1.1.1",
					"1.2.3.4",
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
							{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
		},
		{
			CurrentState: &Endpoint{
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
						Addresses: []corev1.EndpointAddress{},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		var err error

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

		if tc.SetupService != nil {
			if _, err := newResource.k8sClient.CoreV1().Services(tc.SetupService.Namespace).Create(tc.SetupService); err != nil {
				t.Fatalf("%d: error returned setting up k8s service: %s\n", i, err)
			}
		}

		result, err := newResource.newUpdateChange(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatal("case", i, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedCreateState, result) {
			t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedCreateState, result)
		}
	}
}
