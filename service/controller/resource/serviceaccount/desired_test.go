package serviceaccount

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/pkg/label"
)

func Test_Resource_ServiceAccount_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj          interface{}
		ExpectedName string
	}{
		{
			Obj: &v1alpha2.KVMCluster{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						label.Cluster: "al9qy",
					},
					Name: "al9qy",
				},
			},
			ExpectedName: "al9qy",
		},
		{
			Obj: &v1alpha2.KVMCluster{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						label.Cluster: "my-cluster",
					},
					Name: "something-else",
				},
			},
			ExpectedName: "my-cluster",
		},
	}

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

	for i, tc := range testCases {
		result, err := newResource.GetDesiredState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		name := result.(*corev1.ServiceAccount).Name
		if tc.ExpectedName != name {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedName, name)
		}
	}
}
