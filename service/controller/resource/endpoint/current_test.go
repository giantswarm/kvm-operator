package endpoint

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	g8sfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoint_GetCurrentState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		SetupEndpoints    []*corev1.Endpoints
		ExpectedEndpoints interface{}
	}{
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip":      "1.1.1.1",
						"endpoint.kvm.giantswarm.io/service": "TestService",
					},
				},
			},
			ExpectedEndpoints: &Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip": "1.1.1.1",
					},
				},
			},
			ExpectedEndpoints: nil,
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
				},
			},
			ExpectedEndpoints: nil,
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"jabber": "1.1.1.1",
						"wocky":  "TestService",
					},
				},
			},
			ExpectedEndpoints: nil,
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip":      "1.1.1.1",
						"endpoint.kvm.giantswarm.io/service": "TestService",
					},
				},
			},
			SetupEndpoints: []*corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "1.1.1.1",
								},
							},
						},
					},
				},
			},
			ExpectedEndpoints: &Endpoint{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "1.1.1.1",
					},
				},
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip":      "1.1.1.1",
						"endpoint.kvm.giantswarm.io/service": "TestService",
					},
				},
			},
			SetupEndpoints: []*corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []corev1.EndpointSubset{
						{
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
			ExpectedEndpoints: &Endpoint{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "1.1.1.1",
					},
					{
						IP: "1.2.3.4",
					},
				},
				IPs: []string{
					"1.1.1.1",
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"endpoint.kvm.giantswarm.io/ip":      "1.1.1.1",
						"endpoint.kvm.giantswarm.io/service": "TestService",
					},
				},
			},
			SetupEndpoints: []*corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "TestService",
						Namespace: "TestNamespace",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "1.1.1.1",
								},
							},
						},
						{
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
			ExpectedEndpoints: &Endpoint{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "1.1.1.1",
					},
					{
						IP: "1.2.3.4",
					},
				},
				IPs: []string{
					"1.1.1.1",
					"1.2.3.4",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
		},
	}
	var err error

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
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
				if _, err := fakeK8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(k8sEndpoint); err != nil {
					t.Fatalf("%d: error returned setting up k8s endpoint: %s\n", i, err)
				}
			}
			result, err := newResource.GetCurrentState(resourcecanceledcontext.NewContext(context.TODO(), make(chan struct{})), tc.Obj)
			if err != nil {
				t.Fatal("case", i, "expected", nil, "got", err)
			}
			if !reflect.DeepEqual(tc.ExpectedEndpoints, result) {
				t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedEndpoints, result)
			}
		})
	}
}
