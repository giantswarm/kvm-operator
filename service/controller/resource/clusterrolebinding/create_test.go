package clusterrolebinding

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/scheme"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint
)

func Test_Resource_ClusterRoleBinding_newCreateChange(t *testing.T) {
	testCases := []struct {
		Obj                              interface{}
		CurrentState                     interface{}
		DesiredState                     interface{}
		ExpectedClusterRoleBindingsNames []string
	}{
		// Test 1, in case current state and desired state are empty the create
		// state should be empty.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState:                     []*apiv1.ClusterRoleBinding{},
			DesiredState:                     []*apiv1.ClusterRoleBinding{},
			ExpectedClusterRoleBindingsNames: []string{},
		},

		// Test 2, in case current state equals desired state the create state
		// should be empty.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
			},
			DesiredState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
			},
			ExpectedClusterRoleBindingsNames: []string{},
		},

		// Test 3, in case current state misses one item of desired state the create
		// state should contain the missing item of the desired state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{},
			DesiredState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
			},
			ExpectedClusterRoleBindingsNames: []string{
				"cluster-role-binding-1",
			},
		},

		// Test 4, in case current state misses items of desired state the create
		// state should contain the missing items of the desired state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{},
			DesiredState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-2",
					},
				},
			},
			ExpectedClusterRoleBindingsNames: []string{
				"cluster-role-binding-1",
				"cluster-role-binding-2",
			},
		},

		// Test 5, in case current state contains one item not being in desired
		// state the create state should not contain the missing item of the desired
		// state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
			},
			DesiredState:                     []*apiv1.ClusterRoleBinding{},
			ExpectedClusterRoleBindingsNames: []string{},
		},

		// Test 6, in case current state contains items not being in desired state
		// the create state should not contain the missing items of the desired
		// state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-2",
					},
				},
			},
			DesiredState:                     []*apiv1.ClusterRoleBinding{},
			ExpectedClusterRoleBindingsNames: []string{},
		},

		// Test 7, in case current state contains some items of desired state the
		// create state should contain the items being in desired state which are
		// not in create state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-2",
					},
				},
			},
			DesiredState: []*apiv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-3",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-role-binding-4",
					},
				},
			},
			ExpectedClusterRoleBindingsNames: []string{
				"cluster-role-binding-3",
				"cluster-role-binding-4",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := Config{
			CtrlClient: fake.NewFakeClientWithScheme(scheme.Scheme),
			Logger:     microloggertest.New(),

			ClusterRoleGeneral: "test-role",
			ClusterRolePSP:     "test-role-psp",
		}
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.newCreateChange(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		clusterRoleBindings, ok := result.([]*apiv1.ClusterRoleBinding)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ClusterRoleBinding{}, result)
		}

		if len(clusterRoleBindings) != len(tc.ExpectedClusterRoleBindingsNames) {
			t.Fatalf("case %d expected %d cluster role bindings got %d", i+1, len(tc.ExpectedClusterRoleBindingsNames), len(clusterRoleBindings))
		}
	}
}
