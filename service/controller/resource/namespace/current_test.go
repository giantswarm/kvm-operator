package namespace

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint

	"github.com/giantswarm/kvm-operator/pkg/test"
)

func Test_Resource_Namespace_GetCurrentState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		ExpectedNamespace *corev1.Namespace
	}{
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			ExpectedNamespace: nil,
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
		result, err := newResource.GetCurrentState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if !reflect.DeepEqual(tc.ExpectedNamespace, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedNamespace, result)
		}
	}
}
