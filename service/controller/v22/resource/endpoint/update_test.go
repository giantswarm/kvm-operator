package endpoint

import (
	"context"
	"reflect"
	"strconv"
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
			ExpectedCreateState: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
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

			result, err := newResource.newUpdateChange(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
			if err != nil {
				t.Fatal("case", i, "expected", nil, "got", err)
			}
			if !reflect.DeepEqual(tc.ExpectedCreateState, result) {
				t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedCreateState, result)
			}
		})
	}
}
