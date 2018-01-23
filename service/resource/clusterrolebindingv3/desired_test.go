package clusterrolebindingv3

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/rbac/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_CloudConfig_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                         interface{}
		ExpectedClusterRoleBindings []*apiv1.ClusterRoleBinding
	}{
		// Test 1, check it returns a couple of cluster role bindings
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			ExpectedClusterRoleBindings: []*apiv1.ClusterRoleBinding{
				{
					TypeMeta: apismetav1.TypeMeta{
						Kind:       "ClusterRoleBinding",
						APIVersion: apiv1.GroupName,
					},
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "general-cluster-role-binding",
					},
				},
				{
					TypeMeta: apismetav1.TypeMeta{
						Kind:       "ClusterRoleBinding",
						APIVersion: apiv1.GroupName,
					},
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "psp-cluster-role-binding",
					},
				},
			},
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
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		clusterRoleBindings, ok := result.([]*apiv1.ClusterRoleBinding)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ClusterRoleBinding{}, result)
		}

		if len(clusterRoleBindings) != len(tc.ExpectedClusterRoleBindings) {
			t.Fatalf("case %d expected %d cluster role bindings got %d", i+1, len(tc.ExpectedClusterRoleBindings), len(clusterRoleBindings))
		}
	}
}
