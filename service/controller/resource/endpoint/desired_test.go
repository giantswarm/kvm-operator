package endpoint

import (
	"context"
	"reflect"
	"testing"

	g8sfake "github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoint_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                  interface{}
		ExpectedEndpoint     interface{}
		ExpectedErrorHandler func(error) bool
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
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			ExpectedEndpoint: &Endpoint{
				IPs: []string{
					"1.1.1.1",
				},
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			ExpectedErrorHandler: nil,
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
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			ExpectedErrorHandler: IsMissingAnnotationError,
			ExpectedEndpoint:     nil,
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
					Annotations: map[string]string{
						"jabber": "1.1.1.1",
						"wocky":  "abcd",
					},
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
			ExpectedErrorHandler: IsMissingAnnotationError,
			ExpectedEndpoint:     nil,
		},
		{
			Obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestPod",
					Namespace: "TestNamespace",
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
			ExpectedErrorHandler: IsMissingAnnotationError,
			ExpectedEndpoint:     nil,
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

		result, err := newResource.GetDesiredState(resourcecanceledcontext.NewContext(context.TODO(), make(chan struct{})), tc.Obj)
		if err != nil && tc.ExpectedErrorHandler == nil {
			t.Fatalf("case %d unexpected error returned getting desired state: %s\n", i+1, err)
		}
		if err != nil && !tc.ExpectedErrorHandler(err) {
			t.Fatalf("case %d incorrect error returned getting desired state: %s\n", i+1, err)
		}
		if err == nil && tc.ExpectedErrorHandler != nil {
			t.Fatalf("case %d expected error not returned getting desired state\n", i+1)
		}

		if !reflect.DeepEqual(tc.ExpectedEndpoint, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedEndpoint, result)
		}
	}
}
