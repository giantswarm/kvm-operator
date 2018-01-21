package clusterrolebindingv2

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/rbac/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_CloudConfig_newUpdateChange(t *testing.T) {
	testCases := []struct {
		Ctx                                context.Context
		Obj                                interface{}
		CurrentState                       interface{}
		DesiredState                       interface{}
		ExpectedClusterRoleBindinsToUpdate []*apiv1.ClusterRoleBinding
	}{
		// Test 1, in case current state and desired state are empty the update
		// state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState:                       []*apiv1.ClusterRoleBinding{},
			DesiredState:                       []*apiv1.ClusterRoleBinding{},
			ExpectedClusterRoleBindinsToUpdate: nil,
		},

		// Test 2, in case current state and desired state are equal the update
		// state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
					Subjects: []apiv1.Subject{
						{
							Kind:      apiv1.ServiceAccountKind,
							Namespace: "my-cluster",
							Name:      "my-cluster",
						},
					},
					RoleRef: apiv1.RoleRef{
						APIGroup: apiv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "cluster-role",
					},
				},
			},
			DesiredState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
					Subjects: []apiv1.Subject{
						{
							Kind:      apiv1.ServiceAccountKind,
							Namespace: "my-cluster",
							Name:      "my-cluster",
						},
					},
					RoleRef: apiv1.RoleRef{
						APIGroup: apiv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "cluster-role",
					},
				},
			},
			ExpectedClusterRoleBindinsToUpdate: nil,
		},

		// Test 3, in case current state contains two items and desired state is
		// contains the same state but one object is modified internally the update
		// state should contain the the modified item from the current state.
		{
			Ctx: context.TODO(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
					Subjects: []apiv1.Subject{
						{
							Kind:      apiv1.ServiceAccountKind,
							Namespace: "my-cluster",
							Name:      "my-cluster",
						},
					},
					RoleRef: apiv1.RoleRef{
						APIGroup: apiv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "cluster-role-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "cluster-role-binding-2",
					},
					Subjects: []apiv1.Subject{
						{
							Kind:      apiv1.ServiceAccountKind,
							Namespace: "my-cluster",
							Name:      "my-cluster",
						},
					},
					RoleRef: apiv1.RoleRef{
						APIGroup: apiv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "cluster-role-2",
					},
				},
			},
			DesiredState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
					Subjects: []apiv1.Subject{
						{
							Kind:      apiv1.ServiceAccountKind,
							Namespace: "my-cluster",
							Name:      "my-cluster",
						},
					},
					RoleRef: apiv1.RoleRef{
						APIGroup: apiv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "cluster-role-3",
					},
				},
			},
			ExpectedClusterRoleBindinsToUpdate: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
					Subjects: []apiv1.Subject{
						{
							Kind:      apiv1.ServiceAccountKind,
							Namespace: "my-cluster",
							Name:      "my-cluster",
						},
					},
					RoleRef: apiv1.RoleRef{
						APIGroup: apiv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "cluster-role-3",
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
		updateState, err := newResource.newUpdateChange(tc.Ctx, tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		clusterRoleBindingsToUpdate, ok := updateState.([]*apiv1.ClusterRoleBinding)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ClusterRoleBinding{}, updateState)
		}
		if !reflect.DeepEqual(clusterRoleBindingsToUpdate, tc.ExpectedClusterRoleBindinsToUpdate) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedClusterRoleBindinsToUpdate, clusterRoleBindingsToUpdate)
		}
	}
}
