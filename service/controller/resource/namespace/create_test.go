package namespace

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/giantswarm/kvm-operator/pkg/test"
)

func Test_Resource_Namespace_newCreateChange(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		Cur               interface{}
		Des               interface{}
		ExpectedNamespace *corev1.Namespace
	}{
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			Cur: &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
			Des: &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
			ExpectedNamespace: nil,
		},

		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			Cur: nil,
			Des: &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
			ExpectedNamespace: &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
		},
	}

	var newResource *Resource
	{
		resourceConfig := Config{
			CtrlClient: fake.NewFakeClientWithScheme(test.Scheme),
			Logger:     microloggertest.New(),
		}
		var err error
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.newCreateChange(context.TODO(), tc.Obj, tc.Cur, tc.Des)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if tc.ExpectedNamespace == nil {
			if tc.ExpectedNamespace != result {
				t.Fatal("case", i+1, "expected", tc.ExpectedNamespace, "got", result)
			}
		} else {
			name := result.(*corev1.Namespace).Name
			if tc.ExpectedNamespace.Name != name {
				t.Fatal("case", i+1, "expected", tc.ExpectedNamespace.Name, "got", name)
			}
		}
	}
}
